package libtest

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"testing"
	"time"

	"stuartdd.com/lib"
)

var (
	mapData      *lib.DataRoot
	structData   string
	dataFileName = "TestDataTypes.json"
)

/*
go test -v -run TestTreeMapping
map[:[UserA UserB]

UserA:[UserA.pwHints UserA.notes]
UserA.notes:[UserA.notes.note]
UserA.pwHints:[UserA.pwHints.GMailA UserA.pwHints.PrincipalityA]

UserB:[UserB.pwHints UserB.notes]
UserB.notes:[UserB.notes.link UserB.notes.note]
UserB.pwHints:[UserB.pwHints.GMail B UserB.pwHints.Principality B]]
*/
func TestTreeMapping(t *testing.T) {
	loadDataMap(dataFileName)
	assertMapData("", "[UserA UserB]")
	assertMapData("UserA", "[UserA.notes UserA.pwHints]")
	assertMapData("UserA.notes", "[UserA.notes.note]")
	assertMapData("UserA.pwHints", "[UserA.pwHints.GMailA UserA.pwHints.PrincipalityA]")
	assertMapData("UserB", "[UserB.notes UserB.pwHints]")
	assertMapData("UserB.notes", "[UserB.notes.link UserB.notes.note]")
	assertMapData("UserB.pwHints", "[UserB.pwHints.GMail B UserB.pwHints.Principality B]")
}

func assertMapData(id, val string) {
	if fmt.Sprintf("%s", mapData.GetNavIndex(id)) != val {
		log.Fatalf("Nav Map id:%s != %s. It is %s. file:%s\n", id, val, mapData.GetNavIndex(id), dataFileName)
	}
}

/*
go test -v -run TestParseJson
*/
func TestLoadAndParseJson(t *testing.T) {
	loadDataMap(dataFileName)
	testStructContains("1:  :groups")
	testStructContains("2:    :UserA")
	testStructContains("2:    :UserB")
	testStructContains("3:      :pwHints")
	testStructContains("4:        :GMail")
	testStructContains("4:        -note = An")
	testStructContains("5:          -notes = https:")
	if mapData.GetTimeStamp().Format(time.UnixDate) != "Fri Jul 30 21:25:10 BST 2021" {
		log.Fatalf("Timestamp dis not parse to UnixDate correctly. file:%s\n", dataFileName)
	}
}

func loadDataMap(fileName string) {
	if mapData == nil {
		fd := loadTestData(fileName)
		md, err := lib.NewDataRoot(fd)
		if err != nil {
			log.Fatalf("error creating new DataRoot file:%s %v\n", fileName, err)
		}
		_, err = md.ToJson()
		if err != nil {
			log.Fatalf("error in ToJson file:%s %v\n", fileName, err)
		}
		mapData = md
		structData = mapData.ToStruct()
	}
}

func testStructContains(s string) {
	if !strings.Contains(structData, s) {
		log.Fatalf("error missing data: '%s'. file:%s\n", s, dataFileName)
	}
}

func loadTestData(fileName string) []byte {
	dat, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatalf("Unable to read %s, Error: %s", fileName, err)
		return make([]byte, 0)
	}
	return dat
}
