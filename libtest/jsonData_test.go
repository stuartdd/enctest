package libtest

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"stuartdd.com/lib"
)

func TestJsonDataLoad(t *testing.T) {
	dat, err := os.ReadFile("TestDataTypes.json")
	if err != nil {
		fmt.Printf("Failed to read file TestDataTypes.json. Error %s\n", err.Error())
	}
	jd, err := lib.NewJsonData(dat, updateMap)
	if err != nil {
		t.Errorf("Should not have thrown err")
	}

	testNavIndex(t, jd, "", "[Stuart UserA UserB]")
	testNavIndex(t, jd, "Stuart", "[Stuart.notes Stuart.pwHints]")
	testNavIndex(t, jd, "Stuart.pwHints", "[Stuart.pwHints.application]")
	testNavIndex(t, jd, "UserA", "[UserA.notes UserA.pwHints]")
	testNavIndex(t, jd, "UserA.pwHints", "[UserA.pwHints.MyApp UserA.pwHints.PrincipalityA]")
	testNavIndex(t, jd, "UserB", "[UserB.notes UserB.pwHints]")
	testNavIndex(t, jd, "UserB.pwHints", "[UserB.pwHints.GMail B UserB.pwHints.Principality B]")
}

func testNavIndex(t *testing.T, jd *lib.JsonData, name, contains string) {
	ni := jd.GetNavIndex(name)
	if ni == nil {
		t.Errorf("GetNavIndex2(\"%s\") returned nil", name)
	}
	if strings.Contains(fmt.Sprintf("%s", ni), contains) {
		return
	}
	t.Errorf("GetNavIndex2(\"%s\") returned '%s'. Does not contain '%s'", name, ni, contains)
}

func TestJsonDataNoTimeStamp(t *testing.T) {
	_, err := lib.NewJsonData([]byte("{\"text\":\"This is NOT one of those times\"}"), updateMap)
	if err == nil {
		t.Errorf("Should have thrown err. timeStampStr does not exist in data root")
	}
}

func updateMap(a, b, c string, e error) {
	if e == nil {
		fmt.Printf("Updated:%s, %s, %s\n", a, b, c)
	} else {
		fmt.Printf("UpdatedError:%s\n", e.Error())
	}
}
