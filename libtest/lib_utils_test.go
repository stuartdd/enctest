package libtest

import (
	"fmt"
	"testing"

	"github.com/stuartdd2/JsonParser4go/parser"
	"stuartdd.com/lib"
)

var ()

func TestGetPathAfterDataRoot(t *testing.T) {
	p1 := parser.NewBarPath(fmt.Sprintf("%s|b|c", lib.DataMapRootName))
	p1 = lib.GetPathAfterDataRoot(p1)
	if p1.String() != "b|c" {
		t.Errorf("1 Should return path 'b|c' not %s", p1.String())
	}
	p1 = parser.NewBarPath("a|b|c")
	p1 = lib.GetPathAfterDataRoot(p1)
	if p1.String() != "a|b|c" {
		t.Errorf("2 Should return path 'a|b|c' not %s", p1.String())
	}
	p1 = parser.NewBarPath("")
	p1 = lib.GetPathAfterDataRoot(p1)
	if p1.String() != "" {
		t.Errorf("3 Should return path '' not %s", p1.String())
	}
	p1 = parser.NewBarPath(fmt.Sprintf("a|%s|b|c", lib.DataMapRootName))
	p1 = lib.GetPathAfterDataRoot(p1)
	if p1.String() != "b|c" {
		t.Errorf("4 Should return path 'b|c' not %s", p1.String())
	}
	p1 = parser.NewBarPath(fmt.Sprintf("a|b|%s|c", lib.DataMapRootName))
	p1 = lib.GetPathAfterDataRoot(p1)
	if p1.String() != "c" {
		t.Errorf("4 Should return path 'c' not %s", p1.String())
	}
	p1 = parser.NewBarPath(fmt.Sprintf("a|b|c|%s", lib.DataMapRootName))
	p1 = lib.GetPathAfterDataRoot(p1)
	if p1.String() != "" {
		t.Errorf("5 Should return path '' not %s", p1.String())
	}
}
func TestLibToolsPadRight(t *testing.T) {
	if len(lib.PadRight("stu", 5)) != 5 {
		t.Errorf("Should = 5")
	}
	if len(lib.PadRight("stu", 4)) != 4 {
		t.Errorf("Should = 4")
	}
	if len(lib.PadRight("stu", 3)) != 3 {
		t.Errorf("Should = 3")
	}
	if len(lib.PadRight("stu", 2)) != 2 {
		t.Errorf("Should = 2")
	}
	if lib.PadRight("stu", 5) != "stu  " {
		t.Errorf("Should = 'stu  '")
	}
	if lib.PadRight("stu", 4) != "stu " {
		t.Errorf("Should = 'stu '")
	}
	if lib.PadRight("stu", 3) != "stu" {
		t.Errorf("Should = 'stu'")
	}
	if lib.PadRight("stu", 2) != "st" {
		t.Errorf("Should = 'st'")
	}
	if lib.PadRight("stu", 1) != "s" {
		t.Errorf("Should = ' s'")
	}
	if lib.PadRight("stu", 0) != "" {
		t.Errorf("Should = ''")
	}

}

func TestLibToolsPadLeft(t *testing.T) {
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
