tablefmt
====

### Usage

```
$ ./tablefmt -h

Usage: ./tablefmt [OPTIONS] FILE [FILE...]

Format table data

Options:
  -fN-M, --fieldsN-M FORMAT
    format the fields as C printf
      N    N'th field, counted from 1
      N-   from N'th field to end of record
      N-M  from N'th to M'th (included) field
      -M   from first to M'th (included) field
  -t, --type TYPE
    set input table type (csv, tsv, auto)
  -i, --inplace
    edit files in place
  -v, --version
    output version infomation and exit
  -h, --help
    display this help and exit
```

### Example

```
$ cat table.csv
1,2,3
4,5,6
7,8,9

$ tablefmt -f2 %02d table.csv
1,02,3
4,05,6
7,08,9

$ cat table.csv | ./tablefmt -f3 %.2f -
1,2,3.00
4,5,6.00
7,8,9.00
```

