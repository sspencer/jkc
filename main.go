package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"github.com/Jeffail/gabs/v2"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type config struct {
	skipKeys     []string
	skipPartials []string
	skipIds      bool
}

type field struct {
	keyCount    int
	boolCount   int
	numberCount int
	stringCount int
	typeGuess   string // best guess of the type
	dir         string // directory name (hint: name it same as "entity")
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
	var skipKeys, skipPartials arrayFlags
	flag.Var(&skipKeys, "v", "skip this \"key\" (invert match)")                    // can specify more than once
	flag.Var(&skipPartials, "p", "skip \"key\" with this substring (invert match)") // can specify more than once
	skipIds := flag.Bool("i", false, "skip keys that looks like push ids or uuids")
	csvOut := flag.Bool("csv", false, "output in CSV format")
	tsvOut := flag.Bool("tsv", false, "output in TSV format")

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

	countMap := make(map[string]field)

	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	cfg := config{skipKeys: skipKeys, skipIds: *skipIds, skipPartials: skipPartials}
	var jsonFiles []string
	for _, dir := range args {
		if dir[0] != os.PathSeparator {
			dir = filepath.Join(cwd, dir)
		}

		if stat, err := os.Stat(dir); err == nil && stat.IsDir() {
			fileNames, err := walkDir(dir, countMap, cfg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR walking directory %q: %s\n", dir, err)
				os.Exit(1)
			} else {
				jsonFiles = append(jsonFiles, fileNames...)
			}
		} else {
			fmt.Fprintf(os.Stderr, "ERROR %q is not a directory\n", dir)
			os.Exit(1)
		}
	}

	fmt.Println("jsonFIles:", jsonFiles)
	for _, fileName := range jsonFiles {
		if err := countPaths(fileName, countMap, cfg); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}

	out := os.Stdout
	if *csvOut {
		csvPrint(out, countMap, ',')
	} else if *tsvOut {
		csvPrint(out, countMap, '\t')
	} else {
		prettyPrint(out, countMap)
	}
}

// recursively walk directory looking for all *.json files
func walkDir(dir string, countMap map[string]field, cfg config) ([]string, error) {
	fileSystem := os.DirFS(dir)

	var fileNames []string
	err := fs.WalkDir(fileSystem, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}

		if d.Type().IsRegular() && strings.HasSuffix(d.Name(), ".json") {

			fileNames = append(fileNames, filepath.Join(dir, path))
		}

		return nil
	})

	return fileNames, err
}

// read json file, keyCount unique key paths
func countPaths(filePath string, countMap map[string]field, cfg config) error {

	data, err := os.ReadFile(filePath)
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

	dir := filepath.Base(filepath.Dir(filePath))

	for keyPath, val := range flat {
		key := normalizeKeyPath(keyPath, cfg)

		if len(key) > 0 {
			ct := countMap[key]
			ct.keyCount++
			ct.dir = dir

			switch val.(type) {
			case int:
			case float64:
				ct.numberCount++
			case string:
				ct.stringCount++
			case bool:
				ct.boolCount++
			}

			countMap[key] = ct
		}
	}

	return nil
}

// remove array subscripts
//     FROM: pets.1.name
//       TO: pets.name
// while skipping over "skip keys" (returns "")
func normalizeKeyPath(keyPath string, cfg config) string {
	keys := strings.Split(keyPath, ".")
	var newKeys []string
	for _, key := range keys {
		for _, skipKey := range cfg.skipKeys {
			if key == skipKey {
				newKeys = append(newKeys, skipKey)
				newKeys = append(newKeys, "*")
				return strings.Join(newKeys, ".")
			}
		}

		for _, skipPartial := range cfg.skipPartials {
			if strings.Contains(key, skipPartial) {
				newKeys = append(newKeys, "*"+skipPartial+"*")
				return strings.Join(newKeys, ".")
			}
		}

		if cfg.skipIds && looksLikeId(key) {
			newKeys = append(newKeys, "<id>")
		} else if _, err := strconv.Atoi(key); err != nil {
			newKeys = append(newKeys, key)
		} else {
			//newKeys = append(newKeys, "[n]") // was using "[*]" before
			array := "[*]"
			n := len(newKeys)
			if n > 0 {
				newKeys[n-1] = newKeys[n-1] + array
			} else {
				newKeys = append(newKeys, array)
			}
		}
	}

	return strings.Join(newKeys, ".")
}

func looksLikeId(key string) bool {
	n := len(key)
	if n == 28 ||
		strings.HasPrefix(key, "-") ||
		strings.HasPrefix(key, "kpf") ||
		strings.HasPrefix(key, "_test") {
		return true
	} else if n == 36 && strings.Contains(key, "-") {
		return true
	}

	return false
}

// Sort map keys and align keyCount based on longest key path
func prettyPrint(out io.Writer, m map[string]field) {
	var maxLenKey, maxLenType int
	guessTypes(m)
	keys := make([]string, 0, len(m))
	for k, ct := range m {
		if len(k) > maxLenKey {
			maxLenKey = len(k)
		}
		countLen := len(ct.typeGuess)
		if countLen > maxLenType {
			maxLenType = countLen
		}

		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		ct := m[k]
		fmt.Fprintf(out, "%s%s  %s%s  %d\n", k, strings.Repeat(" ", maxLenKey-len(k)), ct.typeGuess, strings.Repeat(" ", maxLenType-len(ct.typeGuess)), ct.keyCount)
	}
}

func csvPrint(out io.Writer, m map[string]field, separator rune) {
	guessTypes(m)
	keys := sortKeys(m)
	w := csv.NewWriter(out)
	w.Comma = separator
	for _, k := range keys {
		ct := m[k]
		if err := w.Write([]string{k, ct.typeGuess, ct.dir, strconv.Itoa(ct.keyCount)}); err != nil {
			log.Fatalln(err)
		}
	}

	w.Flush()
}

func sortKeys(m map[string]field) []string {
	keys := make([]string, 0, len(m))
	for k := range m {

		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func guessTypes(countMap map[string]field) {

	for key, val := range countMap {
		type countType struct {
			typeKey string
			typeCnt int
		}

		var ss []countType
		if val.stringCount > 0 {
			ss = append(ss, countType{"string", val.stringCount})
		}
		if val.numberCount > 0 {
			ss = append(ss, countType{"number", val.numberCount})
		}
		if val.boolCount > 0 {
			ss = append(ss, countType{"boolean", val.boolCount})
		}

		sort.Slice(ss, func(i, j int) bool {
			return ss[i].typeCnt > ss[j].typeCnt
		})

		ct := countMap[key]

		if len(ss) == 0 {
			ct.typeGuess = "unknown"
		} else if len(ss) == 1 {
			ct.typeGuess = ss[0].typeKey
		} else {
			ct.typeGuess = ss[0].typeKey + "(*)"
		}

		countMap[key] = ct
	}
}
