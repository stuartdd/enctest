package libtest

import (
	"fmt"
	"io/ioutil"
	"log"
	"reflect"
	"testing"

	"stuartdd.com/lib"
)

var (
	mapData      map[string]interface{}
	prettyData   string
	dataFileName = "TestDataTypes.json"
	tabdata      = "                                     "
)

/*
go test -v -run TestParseJson
*/
func TestParseJson(t *testing.T) {
	loadDataMap(dataFileName)
	ind := 0
	for k, v := range mapData {
		if (k != "timeStamp") && (k != "groups") {
			log.Fatalf("Expected keys of timeStamp or groups")
		}
		if reflect.ValueOf(v).Kind() != reflect.String {
			fmt.Printf("%d:%s: %s \n", ind, tabdata[:ind*2], k)
			printMap(v.(map[string]interface{}), 1)
		} else {
			fmt.Printf("%d:%s %s = %s\n", ind, tabdata[:ind*2], k, v)
		}
	}
}

func printMap(m map[string]interface{}, ind int) {
	for k, v := range m {
		if reflect.ValueOf(v).Kind() != reflect.String {
			fmt.Printf("%d:%s: %s \n", ind, tabdata[:ind*2], k)
			printMap(v.(map[string]interface{}), ind+1)
		} else {
			fmt.Printf("%d:%s %s = %s\n", ind, tabdata[:ind*2], k, v)
		}
	}
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
