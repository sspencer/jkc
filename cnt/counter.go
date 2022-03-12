package cnt

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/Jeffail/gabs/v2"
)

type Counter struct {
	skipKeys     []string
	skipPartials []string
	skipIds      bool
	counter      counterMap
}

type counterValue struct {
	keyCount    int
	boolCount   int
	numberCount int
	stringCount int
	typeGuess   string // best guess of the type
	dir         string // directory name (hint: name it same as "entity")
}

type counterMap map[string]counterValue

func NewCounter(skipKeys, skipPartials []string, skipIds bool) *Counter {
	return &Counter{
		skipKeys:     skipKeys,
		skipPartials: skipPartials,
		skipIds:      skipIds,
		counter:      make(counterMap),
	}
}

func partialCounter(c *Counter) *Counter {
	return &Counter{
		skipKeys:     c.skipKeys,
		skipPartials: c.skipPartials,
		skipIds:      c.skipIds,
		counter:      make(counterMap),
	}
}

// CountFiles reads JSON files and counts the number of times each json key path appears.a
func (c *Counter) CountFiles(jsonFiles []string) {
	numFiles := len(jsonFiles)

	if numFiles == 0 {
		return
	} else if numFiles >= 120 {
		c.multiCountFiles(jsonFiles)
		return
	}

	// single threaded
	for _, f := range jsonFiles {
		if err := c.countFile(f); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
}

// multiCountFiles breaks input files into equally sized chunks and counts the keypaths concurrently.
func (c *Counter) multiCountFiles(jsonFiles []string) {
	cpus := runtime.NumCPU()
	chunks := splitSlice(jsonFiles, cpus)
	var counters []*Counter

	wg := sync.WaitGroup{}
	for i := 0; i < cpus; i++ {
		counters = append(counters, partialCounter(c))
		wg.Add(1)
		go func(files []string, m *Counter) {
			for _, f := range files {
				if err := m.countFile(f); err != nil {
					fmt.Fprintln(os.Stderr, err)
				}
			}
			wg.Done()
		}(chunks[i], counters[i])
	}
	wg.Wait()

	for i := 0; i < cpus; i++ {
		c.merge(counters[i])
	}
}

func (c *Counter) merge(m *Counter) {
	for k, v := range m.counter {
		if val, ok := c.counter[k]; ok {
			val.keyCount += v.keyCount
			val.boolCount += v.boolCount
			val.numberCount += v.numberCount
			val.stringCount += v.stringCount
			val.dir = v.dir
			c.counter[k] = val
		} else {
			c.counter[k] = v
		}
	}
}

// parse single json file
func (c *Counter) countFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	json, err := gabs.ParseJSON(data)
	if err != nil {
		return err
	}

	flat, err := json.Flatten()
	if err != nil {
		return err
	}

	dir := filepath.Base(filepath.Dir(filePath))

	for keyPath, val := range flat {
		key := c.normalizeKeyPath(keyPath)

		if len(key) > 0 {
			ct := c.counter[key]
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

			c.counter[key] = ct
		}
	}

	return nil
}

// remove array subscripts
//     FROM: pets.1.name
//       TO: pets.name
// while skipping over "skip keys" (returns "")
func (c *Counter) normalizeKeyPath(keyPath string) string {
	keys := strings.Split(keyPath, ".")
	var newKeys []string
	for _, key := range keys {
		for _, skipKey := range c.skipKeys {
			if key == skipKey {
				newKeys = append(newKeys, skipKey)
				newKeys = append(newKeys, "*")
				return strings.Join(newKeys, ".")
			}
		}

		for _, skipPartial := range c.skipPartials {
			if strings.Contains(key, skipPartial) {
				newKeys = append(newKeys, "*"+skipPartial+"*")
				return strings.Join(newKeys, ".")
			}
		}

		if c.skipIds && looksLikeId(key) {
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

// splitSlice splits a slice in `numberOfChunks` slices.
//
// Based on this gist: https://gist.github.com/siscia/988bf4523918345a6a8285b32e685e03
//
func splitSlice(array []string, numberOfChunks int) [][]string {
	if len(array) == 0 {
		return nil
	}

	result := make([][]string, numberOfChunks)

	if numberOfChunks > len(array) {
		for i := 0; i < len(array); i++ {
			result[i] = []string{array[i]}
		}
		return result
	}

	for i := 0; i < numberOfChunks; i++ {

		min := i * len(array) / numberOfChunks
		max := ((i + 1) * len(array)) / numberOfChunks

		result[i] = array[min:max]

	}

	return result
}
