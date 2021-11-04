package libtest

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stuartdd/jsonParserGo/parser"
	"stuartdd.com/lib"
)

func TestJsonDataRename(t *testing.T) {
	jd := dataLoad(t, "TestDataTypes.json")
	testNavIndex(t, jd, "", "[Stuart UserA UserB]")
	testNavIndex(t, jd, "UserA", "[UserA|notes UserA|pwHints]")
	testNavIndex(t, jd, "UserA|pwHints", "[UserA|pwHints|MyApp UserA|pwHints|PrincipalityA]")

	err := jd.Rename(parser.NewBarPath("UserA"), "RenameA")
	if err != nil {
		t.Errorf("Should not have thrown an err")
	}
	ni := jd.GetNavIndex("UserA")
	if ni != nil {
		t.Errorf("Should not have found UserA")
	}
	ni = jd.GetNavIndex("RenameA")
	if ni == nil {
		t.Errorf("Should have found RenameA")
	}
	testNavIndex(t, jd, "", "[RenameA Stuart UserB]")
	testNavIndex(t, jd, "RenameA", "[RenameA|notes RenameA|pwHints]")
	testNavIndex(t, jd, "Stuart", "[Stuart|notes Stuart|pwHints]")
	testNavIndex(t, jd, "Stuart|pwHints", "[Stuart|pwHints|application]")
	testNavIndexNot(t, jd, "UserA")
	testNavIndexNot(t, jd, "UserA|pwHints")
	testNavIndex(t, jd, "UserB", "[UserB|notes UserB|pwHints]")
	testNavIndex(t, jd, "UserB|pwHints", "[UserB|pwHints|GMail.B UserB|pwHints|Principality.B]")
}
func TestJsonDataRemove(t *testing.T) {
	jd := dataLoad(t, "TestDataTypes.json")
	testNavIndex(t, jd, "UserA", "[UserA|notes UserA|pwHints]")
	testNavIndex(t, jd, "UserA|pwHints", "[UserA|pwHints|MyApp UserA|pwHints|PrincipalityA]")
	jd.Remove(parser.NewBarPath("UserA"), 1)
	testNavIndex(t, jd, "", "[Stuart UserB]")
	testNavIndex(t, jd, "Stuart", "[Stuart|notes Stuart|pwHints]")
	testNavIndex(t, jd, "Stuart|pwHints", "[Stuart|pwHints|application]")
	testNavIndexNot(t, jd, "UserA")
	testNavIndexNot(t, jd, "UserA|pwHints")
	testNavIndex(t, jd, "UserB", "[UserB|notes UserB|pwHints]")
	testNavIndex(t, jd, "UserB|pwHints", "[UserB|pwHints|GMail.B UserB|pwHints|Principality.B]")
	jd.Remove(parser.NewBarPath("UserB"), 1)
	testNavIndex(t, jd, "", "[Stuart]")
	testNavIndex(t, jd, "Stuart", "[Stuart|notes Stuart|pwHints]")
	testNavIndex(t, jd, "Stuart|pwHints", "[Stuart|pwHints|application]")
	testNavIndexNot(t, jd, "UserA")
	testNavIndexNot(t, jd, "UserA|pwHints")
	testNavIndexNot(t, jd, "UserB")
	testNavIndexNot(t, jd, "UserB|pwHints")
	err := jd.Remove(parser.NewBarPath("Stuart"), 1)
	if err == nil {
		t.Errorf("Should have thrown an err")
	}
}
func TestJsonDataLoad(t *testing.T) {
	jd := dataLoad(t, "TestDataTypes.json")
	testNavIndex(t, jd, "", "[Stuart UserA UserB]")
	testNavIndex(t, jd, "Stuart", "[Stuart|notes Stuart|pwHints]")
	testNavIndex(t, jd, "Stuart|pwHints", "[Stuart|pwHints|application]")
	testNavIndex(t, jd, "UserA", "[UserA|notes UserA|pwHints]")
	testNavIndex(t, jd, "UserA|pwHints", "[UserA|pwHints|MyApp UserA|pwHints|PrincipalityA]")
	testNavIndex(t, jd, "UserB", "[UserB|notes UserB|pwHints]")
	testNavIndex(t, jd, "UserB|pwHints", "[UserB|pwHints|GMail.B UserB|pwHints|Principality.B]")
}

func dataLoad(t *testing.T, filename string) *lib.JsonData {
	dat, err := os.ReadFile("TestDataTypes.json")
	if err != nil {
		t.Errorf("Failed to read file TestDataTypes.json. Error %s\n", err.Error())
	}
	jd, err := lib.NewJsonData(dat, updateMap)
	if err != nil {
		t.Errorf("Should not have thrown err")
	}
	return jd
}

func testNavIndex(t *testing.T, jd *lib.JsonData, name, contains string) {
	ni := jd.GetNavIndex(name)
	if ni == nil {
		t.Errorf("GetNavIndex(\"%s\") returned nil", name)
	}
	s := fmt.Sprintf("%s", ni)
	if strings.Contains(s, contains) {
		return
	}
	t.Errorf("GetNavIndex(\"%s\") returned '%s'. Does not contain '%s'", name, ni, contains)
}

func testNavIndexNot(t *testing.T, jd *lib.JsonData, name string) {
	if jd.GetNavIndex(name) != nil {
		t.Errorf("GetNavIndex(%s) should be nil", name)
	}
}

func TestJsonDataNoTimeStamp(t *testing.T) {
	_, err := lib.NewJsonData([]byte("{\"text\":\"This is NOT one of those times\"}"), updateMap)
	if err == nil {
		t.Errorf("Should have thrown err. groups does not exist in data root")
	}
	_, err = lib.NewJsonData([]byte("{\"groups\":\"This is NOT one of those times\"}"), updateMap)
	if err == nil {
		t.Errorf("Should have thrown err. groups is not an object node")
	}
	_, err = lib.NewJsonData([]byte("{\"groups\":{\"note\":\"This is NOT one of those times\"}}"), updateMap)
	if err == nil {
		t.Errorf("Should have thrown err. groups is not an object node")
	}
	_, err = lib.NewJsonData([]byte("{\"groups\":{\"note\":\"This is NOT one of those times\"}, \"timeStamp\":\"TS\"}"), updateMap)
	if err == nil {
		t.Errorf("Should have thrown err. groups is not an object node")
	}
}

func updateMap(a, b, c string, e error) {
	// if e == nil {
	// 	fmt.Printf("Updated:%s, %s, %s\n", a, b, c)
	// } else {
	// 	fmt.Printf("UpdatedError:%s\n", e.Error())
	// }
}
