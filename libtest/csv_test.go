package libtest

import (
	"fmt"
	"testing"

	"stuartdd.com/lib"
)

func TestParseFileBasic(t *testing.T) {
	err := lib.ParseFile("testdata.csvt", func(row, col int, val string) error {
		if row == 0 && col == 0 {
			return fmt.Errorf("error 0,0")
		}
		return nil
	})
	if err == nil {
		t.Errorf("err should not be nil for file found")
	}
}
