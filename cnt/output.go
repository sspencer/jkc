package cnt

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"sort"
	"strconv"
	"strings"
)

// CSVPrint prints the counter to a CSV file
func (c *Counter) CSVPrint(out io.Writer, separator rune) {
	c.guessTypes()
	keys := c.sortKeys()
	w := csv.NewWriter(out)
	w.Comma = separator
	for _, k := range keys {
		ct := c.counter[k]
		if err := w.Write([]string{k, ct.typeGuess, ct.dir, strconv.Itoa(ct.keyCount)}); err != nil {
			log.Fatalln(err)
		}
	}

	w.Flush()
}

// PrettyPrint sorts map keys and align keyCount based on longest key path
func (c *Counter) PrettyPrint(out io.Writer) {
	var maxKeyLen, maxTypeLen, maxDirLen int
	dirCnt := make(map[string]int)
	c.guessTypes()
	keys := make([]string, 0, len(c.counter))
	for k, ct := range c.counter {
		if len(k) > maxKeyLen {
			maxKeyLen = len(k)
		}
		countLen := len(ct.typeGuess)
		if countLen > maxTypeLen {
			maxTypeLen = countLen
		}

		dirCnt[ct.dir]++
		dirLen := len(ct.dir)
		if dirLen > maxDirLen {
			maxDirLen = dirLen
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		ct := c.counter[k]
		if len(dirCnt) == 1 {
			fmt.Fprintf(out, "%s%s  %s%s  %d\n",
				k, strings.Repeat(" ", maxKeyLen-len(k)),
				ct.typeGuess, strings.Repeat(" ", maxTypeLen-len(ct.typeGuess)),
				ct.keyCount)
		} else {
			fmt.Fprintf(out, "%s%s  %s%s  %s%s  %d\n",
				k, strings.Repeat(" ", maxKeyLen-len(k)),
				ct.typeGuess, strings.Repeat(" ", maxTypeLen-len(ct.typeGuess)),
				ct.dir, strings.Repeat(" ", maxDirLen-len(ct.dir)),
				ct.keyCount)
		}
	}
}

func (c *Counter) sortKeys() []string {
	keys := make([]string, 0, len(c.counter))
	for k := range c.counter {

		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func (c *Counter) guessTypes() {

	for key, val := range c.counter {
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

		ct := c.counter[key]

		if len(ss) == 0 {
			ct.typeGuess = "unknown"
		} else if len(ss) == 1 {
			ct.typeGuess = ss[0].typeKey
		} else {
			ct.typeGuess = ss[0].typeKey + "(*)"
		}

		c.counter[key] = ct
	}
}
