# JSON Key Counter

JSON Key counter is a command line utility that recursively walks specified 
directories to count unique key paths found in *.json files.

## Features 

1. Specify one or more directories for `jkc` to recursively descend into
2. Array elements are rolled up into `[*]`
3. Keys may be excluded to reduce noise with `-v key` flag
4. IDs (UUID and PushIds) may be rolled up with the `-i` flag which reduces them to `<id>`
5. Partial keys can be ignored with the `-p` flag
 
## Usage 

```
USAGE:
  jkc <dir1> [<dir2> <dir3> ...]
  -i    skip keys that looks like pushids or uuids
  -p value
    	skip "key" with this substring (invert match)  
  -v value
        skip this "key" (invert match)
```

For the included example `testdata/` directory:

```
$ jkc -i -v type testdata
pets[*].<id>.animal.name:   2
pets[*].<id>.animal.type.*: 2
pets[*].<id>.type.*:        2
pets[*].bark:               6
pets[*].meow:               1
pets[*].name:               7
pets[*].speak:              3
pets[*].type.*:             7
```

## Notes

You may skip more than one key:

    jkc -v input -v output dir1 dir2 dir3
