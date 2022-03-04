# JSON Key Counter

JSON Key counter is a command line utility that recursively walks specified 
directories to count unique key paths found in *.json files.

## Features 

1. Specify one or more directories for `jkc` to recursively descend into
2. Array elements are rolled up into `[*]`
3. Keys may be excluded to reduce noise with `-v key` flag
4. IDs (UUID and PushIds) may be rolled up with the `-i` flag which reduces them to `<id>`
5. Partial keys can be ignored with the `-p` flag
6. Reports with directory crawled may be printed with -csv or -tsv flags.

## Usage 

```
USAGE:
  jkc <dir1> [<dir2> <dir3> ...]
  -csv 	    output in CSV format
  -i	    skip keys that looks like push ids or uuids
  -p value 	skip "key" with this substring (invert match)
  -tsv    	output in TSV format
  -v value 	skip this "key" (invert match)```
```

```
$ jkc -i -v type testdata
pets[*].<id>.animal.name  string   2
pets[*].<id>.animal.type  string   2
pets[*].<id>.type         string   2
pets[*].bark              boolean  6
pets[*].meow              boolean  1
pets[*].name              string   7
pets[*].speak             boolean  3
pets[*].type              string   7
```

## Notes

You may skip more than one key:

    jkc -v input -v output dir1 dir2 dir3
