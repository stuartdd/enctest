package libtest

import (
	"fmt"
	"io/ioutil"
	"log"
	"testing"

	"stuartdd.com/lib"
)

var (
	mapData      map[string]interface{}
	prettyData   string
	dataFileName = "TestDataTypes.json"
)

/*
go test -v -run TestParseJson
*/
func TestParseJson(t *testing.T) {
	loadDataMap(dataFileName)
	fmt.Println(string(prettyData))
}

func loadDataMap(fileName string) {
	if mapData == nil {
		md, err := lib.Parse(loadTestData(fileName))
		if err != nil {
			log.Fatalf("error parsing file:%s %v\n", fileName, err)
		}
		mapData = md
		pd, err := lib.PrettyJson(mapData)
		if err != nil {
			log.Fatalf("error prety print file:%s %v\n", fileName, err)
		}
		prettyData = pd
		fmt.Println(string(prettyData))

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
