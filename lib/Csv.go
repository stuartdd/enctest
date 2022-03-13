package lib

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func ParseFileToMap(fileName string, skipRowZero bool, names []string) ([]map[string]string, error) {
	m := make([]map[string]string, 0)
	err := ParseFile(fileName, func(row, col int, s string) error {
		if skipRowZero {
			if row == 0 {
				return nil
			} else {
				row--
			}
		}
		if col >= len(names) {
			return fmt.Errorf("names list. Index out of bounds index[%d]. Range is (0..%d) ", col, len(names)-1)
		}
		cName := strings.TrimSpace(names[col])
		if cName != "" {
			if len(m) < (row + 1) {
				m = append(m, make(map[string]string))
			}
			m[row][cName] = s
		}
		return nil
	})
	return m, err
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
			err := parseLine(fileName, row, s, call)
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

func parseLine(fileName string, row int, line string, call func(int, int, string) error) error {
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
	err := call(row, col, strings.TrimSpace(sb.String()))
	if err != nil {
		return err
	}
	return nil
}
