package main

import (
	"flag"
	"fmt"
	"github.com/Jeffail/gabs/v2"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type cmdParams struct {
	skipKeys []string
	skipIds  bool
}

type arrayFlags []string

func (i *arrayFlags) String() string {
	return "my string representation"
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

// print dotted path for each json key
func main() {
	var skipKeys arrayFlags
	flag.Var(&skipKeys, "v", "skip this \"key\" (invert match)") // can specify more than once
	skipIds := flag.Bool("i", false, "skip keys that looks like pushids or uuids")

	flag.Usage = func() {
		w := flag.CommandLine.Output() // may be os.Stderr - but not necessarily
		fmt.Fprintf(w, "Count all unique JSON key paths in *.json files found in the specified directories\n\n")
		fmt.Fprintf(w, "USAGE:\n  %s <dir1> [<dir2> <dir3> ...]\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		flag.Usage()
		os.Exit(1)
	}

	countMap := make(map[string]int)

	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	opts := cmdParams{skipKeys: skipKeys, skipIds: *skipIds}
	for _, dir := range args {
		if dir[0] != os.PathSeparator {
			dir = filepath.Join(cwd, dir)
		}

		if stat, err := os.Stat(dir); err == nil && stat.IsDir() {
			walkDir(dir, countMap, opts)
		} else {
			fmt.Fprintf(os.Stderr, "ERROR %q is not a directory\n", dir)
			os.Exit(1)
		}
	}

	prettyPrint(countMap)
}

// recursively walk directory looking for all *.json files
func walkDir(dir string, countMap map[string]int, opts cmdParams) {
	fileSystem := os.DirFS(dir)

	fs.WalkDir(fileSystem, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}

		if d.Type().IsRegular() && strings.HasSuffix(d.Name(), ".json") {
			if err = countPaths(filepath.Join(dir, path), countMap, opts); err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		}

		return nil
	})
}

// read json file, count unique key paths
func countPaths(path string, countMap map[string]int, opts cmdParams) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	json, err := gabs.ParseJSON(data)
	if err != nil {
		return err
	}

	//flat, err := json.FlattenIncludeEmpty()
	flat, err := json.Flatten()
	if err != nil {
		return err
	}

	for keyPath, _ := range flat {
		key := normalizeKeyPath(keyPath, opts)
		if len(key) > 0 {
			countMap[key]++
		}
	}

	return nil
}

// remove array subscripts
//     FROM: pets.1.name
//       TO: pets.name
// while skipping over "skip keys" (returns "")
func normalizeKeyPath(keyPath string, opts cmdParams) string {
	keys := strings.Split(keyPath, ".")
	var newKeys []string
	for _, key := range keys {
		for _, skipKey := range opts.skipKeys {
			if key == skipKey {
				return ""
			}
		}

		if opts.skipIds && looksLikeId(key) {
			newKeys = append(newKeys, "<id>")
		} else if _, err := strconv.Atoi(key); err != nil {
			newKeys = append(newKeys, key)
		} else {
			newKeys = append(newKeys, "[*]")
		}
	}

	return strings.Join(newKeys, ".")
}

func looksLikeId(key string) bool {
	n := len(key)
	if n == 20 && strings.HasPrefix(key, "-") {
		return true
	} else if n == 36 && strings.Contains(key, "-") {
		return true
	}

	return false
}

// Sort map keys and align count based on longest key path
func prettyPrint(m map[string]int) {
	var maxLenKey int
	keys := make([]string, 0, len(m))
	for k := range m {
		if len(k) > maxLenKey {
			maxLenKey = len(k)
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := m[k]
		fmt.Printf("%s:%s %d\n", k, strings.Repeat(" ", maxLenKey-len(k)), v)
	}
}
