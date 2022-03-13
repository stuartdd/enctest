package libtest

import (
	"fmt"
	"strings"
	"testing"

	"stuartdd.com/lib"
)

var (
	IMPORT_CSV_MAP_LIST_DEF = []string{"date", "type", "", "", "ref", "out", "in", ""}
	IMPORT_CSV_MAP_LIST_ERR = []string{"type", "", "", "ref", "val", "val", ""}
)

func TestParseSkipRowZero(t *testing.T) {
	m, err := lib.ParseFileToMap("testdata.csvt", true, IMPORT_CSV_MAP_LIST_DEF)
	testErrorNil(t, err, "ParseFileToMap")
	if len(m) != 11 {
		t.Errorf("err list should contain 11 elements. It contains %d", len(m))
	}
	testRow(t, m[0], "out:27.56", "ref:TESCO STORE 3144")
	testRow(t, m[1], "out:36.56", "ref:BT GROUP PLC")
	testRow(t, m[7], "in:2.91", "ref:INTEREST (GROSS)")
	testRow(t, m[9], "in:1473.41", "ref:BT PENSION A/C")
	testRow(t, m[10], "out:120.00", "ref:IDRIS JOHN HANMER")
}

func TestParseNamesLen(t *testing.T) {
	_, err := lib.ParseFileToMap("testdata.csvt", false, IMPORT_CSV_MAP_LIST_ERR)
	testError(t, err, "Index out of bounds")
}

func TestParseFileBasic(t *testing.T) {
	err := lib.ParseFile("testdata.csvt", func(row, col int, val string) error {
		if row == 0 && col == 0 {
			return fmt.Errorf("%d,%d %s", row, col, val)
		}
		return nil
	})
	testError(t, err, "0,0 Transaction Date")
	err2 := lib.ParseFile("testdata.csvt", func(row, col int, val string) error {
		if row == 11 && col == 7 {
			return fmt.Errorf("%d,%d %s", row, col, val)
		}
		return nil
	})
	testError(t, err2, "11,7 4589.37")
	err3 := lib.ParseFile("testdata.csvt", func(row, col int, val string) error {
		if row == 11 && col == 6 {
			return fmt.Errorf("%d,%d %s", row, col, val)
		}
		return nil
	})
	testError(t, err3, "11,6 ")
}

func testError(t *testing.T, err error, txt string) {
	if err == nil {
		t.Errorf("err should not be nil for ParseFile")
		t.FailNow()
	}
	if strings.Contains(err.Error(), txt) {
		return
	}
	t.Errorf("err '%s' should contain '%s'", err.Error(), txt)
}

func testErrorNil(t *testing.T, err error, txt string) {
	if err != nil {
		t.Errorf("err should be nil. %s", txt)
		t.FailNow()
	}
}

func testRow(t *testing.T, m map[string]string, txt1, txt2 string) {
	s := fmt.Sprintf("%s", m)
	if strings.Contains(s, txt1) && strings.Contains(s, txt2) {
		return
	}
	t.Errorf("err map text '%s' should contain '%s' and '%s'", s, txt1, txt2)
}
