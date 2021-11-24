package main

import (
	"fmt"
	"github.com/Jeffail/gabs/v2"
	"io/fs"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
)

// print dotted path for each json key
func main() {
	countMap := make(map[string]int)

	root, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	fileSystem := os.DirFS(root)

	fs.WalkDir(fileSystem, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Fatal(err)
		}

		if d.Type().IsRegular() && strings.HasSuffix(d.Name(), ".json") {
			if err = countPaths(countMap, path); err != nil {
				log.Fatal(err)
			}
		}

		return nil
	})

	prettyPrint(countMap)
}

func countPaths(countMap map[string]int, path string) error {
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