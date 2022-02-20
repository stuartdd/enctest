package libtest

import (
	"testing"

	"stuartdd.com/lib"
)

var ()

func TestLibTools(t *testing.T) {
	if len(lib.PadLeft("stu", 5)) != 5 {
		t.Errorf("Should = 5")
	}
	if len(lib.PadLeft("stu", 4)) != 4 {
		t.Errorf("Should = 4")
	}
	if len(lib.PadLeft("stu", 3)) != 3 {
		t.Errorf("Should = 3")
	}
	if len(lib.PadLeft("stu", 2)) != 2 {
		t.Errorf("Should = 2")
	}
	if lib.PadLeft("stu", 5) != "  stu" {
		t.Errorf("Should = '  stu'")
	}
	if lib.PadLeft("stu", 4) != " stu" {
		t.Errorf("Should = ' stu'")
	}
	if lib.PadLeft("stu", 3) != "stu" {
		t.Errorf("Should = ' stu'")
	}
	if lib.PadLeft("stu", 2) != "st" {
		t.Errorf("Should = ' st'")
	}
	if lib.PadLeft("stu", 1) != "s" {
		t.Errorf("Should = ' s'")
	}
	if lib.PadLeft("stu", 0) != "" {
		t.Errorf("Should = ''")
	}

}
