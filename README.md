# JSON Key Counter

JSON Key counter is a command line utility that recursively walks all the specified 
directories to count all the unique key paths found in *.json files.

## Features 

1. One or more directories can be specified on the command line and `jkc` will recusively descend into each sub directory
2. Array elements are rolled up into `[*]`
3. Keys may be excluded to reduce noise with `-v key` flag
4. IDs (UUID and PushIds) may be rolled up with the `-i` flag which reduces them to `<id>`

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

For the included example `data/` directory:

```
$ jkc -i -v type data
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
