package main

import (
	"fmt"
	"github.com/Jeffail/gabs/v2"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// print dotted path for each json key
func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Count all unique JSON key paths in *.json files found in the specified directories\n\n")
		fmt.Fprintf(os.Stderr, "USAGE:\n    jkc <dir1> [<dir2> <dir3> ...]\n")
		os.Exit(1)
	}

	countMap := make(map[string]int)

	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	for _, dir := range os.Args[1:] {
		if dir[0] != os.PathSeparator {
			dir = filepath.Join(cwd, dir)
		}

		if stat, err := os.Stat(dir); err == nil && stat.IsDir() {
			walkDir(dir, countMap)
		} else {
			fmt.Fprintf(os.Stderr, "ERROR %q is not a directory\n", dir)
			os.Exit(1)
		}
	}

	prettyPrint(countMap)
}

// recursively walk directory looking for all *.json files
func walkDir(dir string, countMap map[string]int) {
	fileSystem := os.DirFS(dir)

	fs.WalkDir(fileSystem, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}

		if d.Type().IsRegular() && strings.HasSuffix(d.Name(), ".json") {
			if err = countPaths(filepath.Join(dir, path), countMap); err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		}

		return nil
	})
}

// read json file, count unique key paths
func countPaths(path string, countMap map[string]int) error {
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
		countMap[normalizeKeyPath(keyPath)]++
	}

	return nil
}

// remove array subscripts
//     FROM: pets.1.name
//       TO: pets.name
func normalizeKeyPath(keyPath string) string {
	keys := strings.Split(keyPath, ".")
	var newKeys []string
	for _, key := range keys {
		if _, err := strconv.Atoi(key); err != nil {
			newKeys = append(newKeys, key)
		}
	}

	return strings.Join(newKeys, ".")
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
