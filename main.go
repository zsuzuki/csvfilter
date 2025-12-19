package main

import (
	"bufio"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
)

type options struct {
	filePath    string
	filterCol   string
	filterValue string
	sortCol     string
	sortType    string
}

type sortMode int

const (
	sortModeAuto sortMode = iota
	sortModeNum
	sortModeStr
)

func main() {
	opts := parseFlags()

	records, err := readCSV(opts.filePath)
	if err != nil {
		exitWithError(err)
	}
	if len(records) == 0 {
		return
	}

	headers := records[0]
	colIndex := indexByHeader(headers)

	filtered := records
	if opts.filterCol != "" || opts.filterValue != "" {
		filtered, err = applyFilter(records, colIndex, opts)
		if err != nil {
			exitWithError(err)
		}
	}

	if opts.sortCol != "" {
		filtered, err = applySort(filtered, colIndex, opts)
		if err != nil {
			exitWithError(err)
		}
	}

	if err := writeCSV(filtered, os.Stdout); err != nil {
		exitWithError(err)
	}
}

func parseFlags() options {
	var opts options
	flag.StringVar(&opts.filePath, "file", "", "CSV file path (optional, otherwise first arg or stdin)")
	flag.StringVar(&opts.filterCol, "filter", "", "column name for filtering")
	flag.StringVar(&opts.filterValue, "value", "", "substring to match for filtering")
	flag.StringVar(&opts.sortCol, "sort", "", "column name for sorting")
	flag.StringVar(&opts.sortType, "type", "asc", "sort direction: asc/desc or lt/le/gt/ge, optionally :num or :str (e.g. asc:num)")
	flag.Parse()

	if opts.filePath == "" {
		args := flag.Args()
		if len(args) > 0 {
			opts.filePath = args[0]
		}
	}

	return opts
}

func readCSV(filePath string) ([][]string, error) {
	var r io.Reader
	if filePath != "" {
		f, err := os.Open(filePath)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		r = f
	} else {
		r = bufio.NewReader(os.Stdin)
	}

	csvr := csv.NewReader(r)
	return csvr.ReadAll()
}

func writeCSV(records [][]string, w io.Writer) error {
	csvw := csv.NewWriter(w)
	if err := csvw.WriteAll(records); err != nil {
		return err
	}
	csvw.Flush()
	return csvw.Error()
}

func indexByHeader(headers []string) map[string]int {
	m := make(map[string]int, len(headers))
	for i, h := range headers {
		m[h] = i
	}
	return m
}

func applyFilter(records [][]string, colIndex map[string]int, opts options) ([][]string, error) {
	if opts.filterCol == "" || opts.filterValue == "" {
		return nil, errors.New("both -filter and -value must be specified")
	}
	idx, ok := colIndex[opts.filterCol]
	if !ok {
		return nil, fmt.Errorf("filter column not found: %s", opts.filterCol)
	}

	filtered := make([][]string, 0, len(records))
	filtered = append(filtered, records[0])

	for _, row := range records[1:] {
		if idx >= len(row) {
			continue
		}
		if strings.Contains(row[idx], opts.filterValue) {
			filtered = append(filtered, row)
		}
	}

	return filtered, nil
}

func applySort(records [][]string, colIndex map[string]int, opts options) ([][]string, error) {
	if len(records) <= 1 {
		return records, nil
	}
	idx, ok := colIndex[opts.sortCol]
	if !ok {
		return nil, fmt.Errorf("sort column not found: %s", opts.sortCol)
	}

	direction, mode, err := normalizeSortType(opts.sortType)
	if err != nil {
		return nil, err
	}
	if direction == 0 {
		return records, nil
	}

	if mode == sortModeNum {
		if err := validateNumeric(records[1:], idx); err != nil {
			return nil, err
		}
	}

	headers := records[0]
	rows := append([][]string(nil), records[1:]...)

	sort.SliceStable(rows, func(i, j int) bool {
		ai := valueAt(rows[i], idx)
		aj := valueAt(rows[j], idx)

		if mode != sortModeStr {
			if fi, okI := parseFloat(ai); okI {
				if fj, okJ := parseFloat(aj); okJ {
					if direction > 0 {
						return fi < fj
					}
					return fi > fj
				}
			}
		}

		if direction > 0 {
			return ai < aj
		}
		return ai > aj
	})

	return append([][]string{headers}, rows...), nil
}

func normalizeSortType(t string) (int, sortMode, error) {
	trimmed := strings.ToLower(strings.TrimSpace(t))
	base := trimmed
	mode := sortModeAuto
	if parts := strings.SplitN(trimmed, ":", 2); len(parts) == 2 {
		base = strings.TrimSpace(parts[0])
		switch strings.TrimSpace(parts[1]) {
		case "", "auto":
			mode = sortModeAuto
		case "num", "number", "numeric":
			mode = sortModeNum
		case "str", "string", "text":
			mode = sortModeStr
		default:
			return 0, sortModeAuto, fmt.Errorf("unsupported sort mode: %s", parts[1])
		}
	}

	switch base {
	case "", "asc", "lt", "le":
		return 1, mode, nil
	case "desc", "gt", "ge":
		return -1, mode, nil
	default:
		return 0, sortModeAuto, fmt.Errorf("unsupported sort type: %s", t)
	}
}

func valueAt(row []string, idx int) string {
	if idx < len(row) {
		return row[idx]
	}
	return ""
}

func parseFloat(s string) (float64, bool) {
	f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return 0, false
	}
	return f, true
}

func validateNumeric(rows [][]string, idx int) error {
	for _, row := range rows {
		v := strings.TrimSpace(valueAt(row, idx))
		if v == "" {
			return errors.New("numeric sort requested but empty value found")
		}
		if _, ok := parseFloat(v); !ok {
			return fmt.Errorf("numeric sort requested but non-numeric value found: %s", v)
		}
	}
	return nil
}

func exitWithError(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
