package libtest

import (
	"testing"

	"stuartdd.com/lib"
)

func TestNew(t *testing.T) {
	lb := lib.NewLine(20)
	if lb.String() != "" {
		t.Errorf("Buffer [%s] should == ''", lb.String())
	}
	lb.Apply("HI", 4)
	lb.Apply("Low", 4)
	if lb.String() != "HI Low" {
		t.Errorf("Buffer [%s] should == 'HI Low'", lb.String())
	}
	lb.Apply("", 4)
	if lb.String() != "HI Low    " {
		t.Errorf("Buffer [%s] should == 'HI Low    '", lb.String())
	}
	lb.Clear()
	if lb.String() != "" {
		t.Errorf("Buffer [%s] should == ''", lb.String())
	}
}
