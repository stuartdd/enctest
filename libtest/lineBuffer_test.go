package libtest

import (
	"testing"

	"stuartdd.com/lib"
)

func TestNew(t *testing.T) {
	lb := lib.NewLine(18)
	if lb.String() != "" {
		t.Errorf("Buffer [%s] should == ''", lb.String())
	}
	lb.Apply("HI", 4)
	lb.Apply("Low", 4)
	if lb.String() != "HI  Low " {
		t.Errorf("Buffer [%s] should == 'HI  Low '", lb.String())
	}
	lb.Apply("", 4)
	if lb.String() != "HI  Low     " {
		t.Errorf("Buffer [%s] should == 'HI  Low     '", lb.String())
	}
	lb.Clear()
	if lb.String() != "" {
		t.Errorf("Buffer [%s] should == ''", lb.String())
	}
	lb.Apply("aa", 4)
	if lb.String() != "aa  " {
		t.Errorf("Buffer [%s] should == 'aa  '", lb.String())
	}
	lb.Apply("", 4)
	if lb.String() != "aa      " {
		t.Errorf("Buffer [%s] should == 'aa      '", lb.String())
	}

	lb.Clear()
	if lb.String() != "" {
		t.Errorf("Buffer [%s] should == ''", lb.String())
	}
	lb.Apply("....", 4)
	if lb.String() != "...." {
		t.Errorf("Buffer [%s] should == '....'", lb.String())
	}
	lb.ApplyRev("abcd", 6)
	if lb.String() != "....  abcd" {
		t.Errorf("Buffer [%s] should == '....  abcd'", lb.String())
	}
	lb.ApplyRev("1234", 6)
	if lb.String() != "....  abcd  1234" {
		t.Errorf("Buffer [%s] should == '....  abcd  1234'", lb.String())
	}
	lb.Apply("wxyz", 6)
	if lb.String() != "....  abcd  1234wx" {
		t.Errorf("Buffer [%s] should == '....  abcd  1234wx'", lb.String())
	}
	lb.Apply("1234", 4)
	if lb.String() != "....  abcd  1234wx" {
		t.Errorf("Buffer [%s] should == '....  abcd  1234wx'", lb.String())
	}

	lb.Clear()
	lb.Apply("....  abcd  1234", 16)
	if lb.String() != "....  abcd  1234" {
		t.Errorf("Buffer [%s] should == '....  abcd  1234'", lb.String())
	}
	lb.ApplyRev("wxyz", 5)
	if lb.String() != "....  abcd  1234 w" {
		t.Errorf("Buffer [%s] should == '....  abcd  1234 w'", lb.String())
	}

}
