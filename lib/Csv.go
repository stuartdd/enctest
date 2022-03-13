package lib

import (
	"bufio"
	"os"
	"strings"
)

func ParseCsv(fileName string, skipFirstLine bool, dtFormat string, mapList []string) ([]map[string]string, error) {
	d := make([]map[string]string, 0)

	return d, nil
}

func ParseFile(fileName string, call func(int, int, string) error) error {
	f, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	row := 0
	for scanner.Scan() {
		s := scanner.Text()
		s = strings.TrimSpace(s)
		if len(s) > 1 {
			err := ParseLine(fileName, row, s, call)
			if err != nil {
				return err
			}
			row++
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func ParseLine(fileName string, row int, line string, call func(int, int, string) error) error {
	col := 0
	var sb strings.Builder
	for _, c := range line {
		if c == ',' {
			err := call(row, col, strings.TrimSpace(sb.String()))
			if err != nil {
				return err
			}
			col++
			sb.Reset()
		} else {
			sb.WriteRune(c)
		}
	}
	return nil
}
