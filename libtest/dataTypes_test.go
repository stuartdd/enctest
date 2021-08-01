package libtest

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"testing"

	"stuartdd.com/lib"
)

var (
	mapData      *lib.DataRoot
	prettyData   string
	structData   string
	dataFileName = "TestDataTypes.json"
)

/*
go test -v -run TestParseJson
*/
func TestParseJson(t *testing.T) {
	loadDataMap(dataFileName)
	testStructContains("1:  :groups")
	testStructContains("2:    :UserA")
	testStructContains("2:    :UserB")
	testStructContains("3:      :pwHints")
	testStructContains("4:        :GMail")
	testStructContains("4:        -note = An")
	testStructContains("5:          -notes = https:")
	fmt.Println(structData)
}

func loadDataMap(fileName string) {
	if mapData == nil {
		md, err := lib.Parse(loadTestData(fileName))
		if err != nil {
			log.Fatalf("error parsing file:%s %v\n", fileName, err)
		}
		mapData, err = lib.NewDataRoot(md)
		if err != nil {
			log.Fatalf("error creating DataRoot file:%s %v\n", fileName, err)
		}
		pd, err := mapData.ToJson()
		if err != nil {
			log.Fatalf("error pretty print file:%s %v\n", fileName, err)
		}
		structData = mapData.ToStruct()
		prettyData = pd
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
