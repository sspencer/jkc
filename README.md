# JSON Key Counter

JSON Key counter is a command line utility that recursively walks all the specified 
directories to count all the unique key paths found in *.json files.

```
USAGE:

jkc dir1 [dir2 dir3...]
```

For the included example `data/` directory:

```
$ jkc data
pets.bark:     6
pets.meow:     1
pets.name:     7
pets.speak:    3
pets.type:     7
```
