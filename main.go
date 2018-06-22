package main

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"golang.org/x/crypto/ssh/terminal"
)

func ShowUsage(output io.Writer) {
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Usage: "+os.Args[0]+" [OPTIONS] FILE [FILE...]")
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Format table data")
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Options:")
	fmt.Fprintln(output, "  -cN FORMAT")
	fmt.Fprintln(output, "    format the column N as C printf")
	fmt.Fprintln(output, "  -t TYPE")
	fmt.Fprintln(output, "    set input table type (csv, tsv, auto)")
	fmt.Fprintln(output, "  -i")
	fmt.Fprintln(output, "    edit files in place")
	fmt.Fprintln(output, "  -v")
	fmt.Fprintln(output, "    output version infomation and exit")
	fmt.Fprintln(output, "  -h")
	fmt.Fprintln(output, "    display this help and exit")
}

func ShowVersion(output io.Writer) {
	fmt.Fprintln(output, "v1.0.0")
}

const maxLineCount = 1

func DetectTableType(src io.Reader) (string, io.Reader, error) {
	rbuf := make([]byte, 512)

	buf := bytes.NewBuffer(make([]byte, 0, 512))
	for line := 0; line < maxLineCount; {
		n, err := src.Read(rbuf)
		if n > 0 {
			buf.Write(rbuf[:n])

			for i := 0; i < n; i++ {
				if rbuf[i] == '\n' {
					line++
				}
			}
		}

		if err == io.EOF {
			break
		}

		if err != nil {
			return "", nil, err
		}
	}

	tableTypes := map[rune]string{
		',':  "csv",
		'\t': "tsv",
	}

	fields := map[string]int{
		"csv": 0,
		"tsv": 0,
	}

	for delimiter, format := range tableTypes {
		reader := csv.NewReader(bytes.NewReader(buf.Bytes()))
		reader.Comma = delimiter
		reader.LazyQuotes = true
		if records, err := reader.ReadAll(); err == nil {
			for _, record := range records {
				fields[format] += len(record)
			}
		}
	}

	tableType := ""
	max := 0
	for f, count := range fields {
		if count > max {
			tableType = f
			max = count
		}
	}

	if tableType == "" {
		return "", nil, errors.New("unknown format")
	}

	return tableType, io.MultiReader(bytes.NewReader(buf.Bytes()), src), nil
}

func FormatField(format string, field string) (string, error) {
	cmd := exec.Command("printf", format, field)
	stderr := bytes.NewBuffer(make([]byte, 0, 256))
	cmd.Stderr = stderr
	output, err := cmd.Output()
	if err != nil {
		return field, errors.New(stderr.String())
	}
	return string(bytes.TrimSpace(output)), nil
}

func FormatTable(dst io.Writer, src io.Reader, tableType string, columns map[int]string) error {
	var err error

	if tableType == "auto" {
		tableType, src, err = DetectTableType(src)
		if err != nil {
			return err
		}
	}

	reader := csv.NewReader(src)
	reader.LazyQuotes = true
	switch tableType {
	case "tsv":
		reader.Comma = '\t'

	case "csv":
		reader.Comma = ','

	default:
		return errors.New("unknown table type: " + tableType)
	}

	writer := csv.NewWriter(dst)
	if tableType == "tsv" {
		writer.Comma = '\t'
	}

	for {
		record, err := reader.Read()
		if record != nil {
			formated := make([]string, 0, len(record))

			for i, field := range record {
				format, ok := columns[i]
				if ok {
					field, err = FormatField(format, field)
					if err != nil {
						return err
					}
				}
				formated = append(formated, field)
			}

			if err := writer.Write(formated); err != nil {
				return err
			}

			writer.Flush()
		}

		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	var err error

	args := os.Args[1:]
	version := false
	help := false
	columns := map[int]string{}
	tableType := "auto"
	inplace := false
	files := []string{}

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if arg == "-" || !strings.HasPrefix(arg, "-") {
			files = append(files, arg)
			continue
		}

		if arg == "-v" {
			version = true
			continue
		}

		if arg == "-h" {
			help = true
			continue
		}

		if arg == "-i" {
			inplace = true
			continue
		}

		if arg == "-t" {
			if i+1 < len(args) {
				i++
				tableType = args[i]
			}
			continue
		}

		if prefix := arg[:len(arg)-1]; prefix == "-c" {
			column, err := strconv.Atoi(arg[len(prefix):])
			if err != nil {
				continue
			}

			if i+1 < len(args) {
				i++
				columns[column] = args[i]
			}
			continue
		}

		fmt.Fprintln(os.Stderr, "unknown option: "+arg)
		ShowUsage(os.Stderr)
		os.Exit(1)
	}

	if version {
		ShowVersion(os.Stdout)
		return
	}

	if help || len(files) == 0 {
		ShowUsage(os.Stdout)
		return
	}

	for _, file := range files {
		var src io.ReadCloser
		var dst io.Writer

		switch {
		case file == "-":
			if terminal.IsTerminal(int(os.Stdin.Fd())) {
				fmt.Fprintln(os.Stderr, "stdin is not pipe")
				os.Exit(1)
			}
			src = os.Stdin
			dst = os.Stdout

		case inplace:
			src, err = os.Open(file)
			if err != nil {
				fmt.Fprintln(os.Stderr, err.Error()+":"+file)
				os.Exit(1)
			}
			dst = bytes.NewBuffer(make([]byte, 0, 1024))

		default:
			src, err = os.Open(file)
			if err != nil {
				fmt.Fprintln(os.Stderr, err.Error()+":"+file)
				os.Exit(1)
			}
			dst = os.Stdout
		}

		err = FormatTable(dst, src, tableType, columns)
		src.Close()
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error()+":"+file)
			os.Exit(1)
		}

		if buf, ok := dst.(*bytes.Buffer); ok {
			if err := ioutil.WriteFile(file, buf.Bytes(), 0644); err != nil {
				fmt.Fprintln(os.Stderr, err.Error()+":"+file)
				os.Exit(1)
			}
		}
	}
}
