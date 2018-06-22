tablefmt
====

### Usage

```
$ tablefmt -h

Usage: tablefmt [OPTIONS] FILE [FILE...]

Format table data

Options:
  -cN FORMAT
    format the column N as C printf
  -t TYPE
    set input table type (csv, tsv, auto)
  -i
    edit files in place
  -v
    output version infomation and exit
  -h
    display this help and exit
```

### Example

```
$ cat table.csv
1,2,3
4,5,6
7,8,9

$ tablefmt -c1 %02d table.csv
1,02,3
4,05,6
7,08,9

$ cat table.csv | ./tablefmt -c2 %.2f -
1,2,3.00
4,5,6.00
7,8,9.00
```

