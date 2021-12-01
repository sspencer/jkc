# JSON Key Counter

JSON Key counter is a command line utility that recursively walks all the specified 
directories to count all the unique key paths found in *.json files.

```
USAGE:
  jkc <dir1> [<dir2> <dir3> ...]
  -i    skip keys that looks like pushids or uuids
  -v value
        skip this "key" (invert match)
```

For the included example `data/` directory:

```
$ jkc -i -v type data
pets.[*].<id>.animal.name: 2
pets.[*].bark:             6
pets.[*].meow:             1
pets.[*].name:             7
pets.[*].speak:            3
```

**NOTE**

You may skip more than one key:

    jkc -v input -v output dir1 dir2 dir3
