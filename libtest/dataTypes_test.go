package libtest

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"testing"

	"github.com/stuartdd/jsonParserGo/parser"
	"stuartdd.com/lib"
)

var (
	mapData      *lib.JsonData
	structData   string
	dataFileName = "TestDataUnEncrypted.json"
)

/*
go test -v -run TestTreeMapping
map[:[UserA UserB]

UserA:[UserA|pwHints UserA|notes]
UserA|notes:[UserA|notes|note]
UserA|pwHints:[UserA|pwHints.GMailA UserA|pwHints|PrincipalityA]

UserB:[UserB|pwHints UserB|notes]
UserB|notes:[UserB|notes|link UserB|notes.note]
UserB|pwHints:[UserB|pwHints|GMail B UserB|pwHints|Principality B]]
*/
func TestTreeMapping(t *testing.T) {
	loadDataMap(dataFileName)
	assertMapData("", "[Stuart UserA UserB]")
	assertMapData("UserA", "[UserA|notes UserA|pwHints]")
	assertMapData("UserA|notes", "[]")
	assertMapData("UserA|pwHints", "[UserA|pwHints|MyApp UserA|pwHints|PrincipalityA]")
	assertMapData("UserB", "[UserB|notes UserB|pwHints]")
	assertMapData("UserB|notes", "[]")
	assertMapData("UserB|pwHints", "[UserB|pwHints|GMail B UserB|pwHints|Principality B]")
}

func assertMapData(id, val string) {
	s := mapData.GetNavIndex(id)
	if fmt.Sprintf("%s", s) != val {
		log.Fatalf("Nav Map id:%s != %s. It is %s. file:%s\n", id, val, s, dataFileName)
	}
}

func TestGetLastId(t *testing.T) {
	s := lib.GetLastId("")
	if s != "" {
		log.Fatal("1. GetLastId fail. should return empty string")
	}
	s = lib.GetLastId("abc")
	if s != "abc" {
		log.Fatal("1. GetLastId fail. should return 'abc'")
	}
	s = lib.GetLastId("|abc")
	if s != "abc" {
		log.Fatal("1. GetLastId fail. should return 'abc'")
	}
	s = lib.GetLastId("|abc|")
	if s != "" {
		log.Fatal("1. GetLastId fail. should return empty string")
	}
	s = lib.GetLastId("|abc|x")
	if s != "x" {
		log.Fatal("1. GetLastId fail. should 'x'")
	}
	s = lib.GetLastId("|abc| x")
	if s != " x" {
		log.Fatal("1. GetLastId fail. should ' x'")
	}
}

func TestGetDataForSelectedId(t *testing.T) {
	loadDataMap(dataFileName)
	m, err := lib.GetUserDataForUid(mapData.GetDataRoot(), parser.NewBarPath("UserA"))
	if err != nil {
		log.Fatalf("1. GetMapForUid fail. should not return error: '%s'", err.Error())
	}
	if m.GetName() != "UserA" {
		log.Fatalf("1. GetMapForUid fail. should return UserA. actual: '%s'", toJson(m))
	}
	m, err = lib.GetUserDataForUid(mapData.GetDataRoot(), parser.NewBarPath("UserA|pwHints"))
	if err != nil {
		log.Fatalf("1. GetMapForUid fail. should not return error: '%s'", err.Error())
	}
	if m.GetName() != "pwHints" {
		log.Fatalf("1. GetMapForUid fail. should return pwHints. actual: '%s'", toJson(m))
	}
	m, err = lib.GetUserDataForUid(mapData.GetDataRoot(), parser.NewBarPath("UserA|pwHints|MyApp"))
	if err != nil {
		log.Fatalf("1. GetMapForUid fail. should not return error: '%s'", err.Error())
	}
	if m.GetName() != "MyApp" {
		log.Fatalf("1. GetMapForUid fail. should return MyApp. actual: '%s'", toJson(m))
	}
	m, err = lib.GetUserDataForUid(mapData.GetDataRoot(), parser.NewBarPath("UserB|pwHints|GMail B"))
	if err != nil {
		log.Fatalf("1. GetMapForUid fail. should not return error: '%s'", err.Error())
	}
	if m.GetName() != "GMail B" {
		log.Fatalf("1. GetMapForUid fail. should return GMail B. actual: '%s'", toJson(m))
	}
}

func TestGetParentId(t *testing.T) {
	if lib.GetParentId("user") != "user" {
		log.Fatal("1. GetParentId fail. should return 'user'")
	}
	if lib.GetParentId("user|id") != "user" {
		log.Fatal("2. GetParentId fail. should return 'user'")
	}
	if lib.GetParentId("user|id|tew") != "user|id" {
		log.Fatal("3. GetParentId fail. should return 'user|id'")
	}
	if lib.GetParentId("user|id|tew|uuu") != "user|id|tew" {
		log.Fatal("4. GetParentId fail. should return 'user|id|tew'")
	}
	if lib.GetParentId("user id|tew|uuu") != "user id|tew" {
		log.Fatal("4. GetParentId fail. should return 'user id|tew'")
	}
}
func TestGetPathElementAt(t *testing.T) {
	if lib.GetPathElementAt("user", 0) != "user" {
		log.Fatal("1. GetPathElementAt fail. should return 'user'")
	}
	if lib.GetPathElementAt("user", 1) != "" {
		log.Fatal("2. GetPathElementAt fail. should return ''")
	}
	if lib.GetPathElementAt("user", -1) != "" {
		log.Fatal("3. GetPathElementAt fail. should return ''")
	}
	if lib.GetPathElementAt("user|abc", 1) != "abc" {
		log.Fatal("4. GetPathElementAt fail. should return 'abc'")
	}
	if lib.GetPathElementAt("user|abc", 0) != "user" {
		log.Fatal("5. GetPathElementAt fail. should return 'user'")
	}
	if lib.GetPathElementAt("user|abc", -1) != "" {
		log.Fatal("6. GetPathElementAt fail. should return ''")
	}
	if lib.GetPathElementAt("user|abc", 2) != "" {
		log.Fatal("7. GetPathElementAt fail. should return ''")
	}
	if lib.GetPathElementAt("user|abc|1", 2) != "1" {
		log.Fatal("8. GetPathElementAt fail. should return '1'")
	}
	if lib.GetPathElementAt("user|abc|1", 3) != "" {
		log.Fatal("9. GetPathElementAt fail. should return ''")
	}
	if lib.GetPathElementAt("user|abc|1", 99) != "" {
		log.Fatal("10. GetPathElementAt fail. should return ''")
	}
}

func TestGetUserFromPath(t *testing.T) {
	if lib.GetUserFromUid("user") != "user" {
		log.Fatal("GetUserFromPath fail. should return 'user'")
	}
	if lib.GetUserFromUid("user|a") != "user" {
		log.Fatal("GetUserFromPath fail. should return 'user'")
	}
	if lib.GetUserFromUid("user|a|b|c") != "user" {
		log.Fatal("GetUserFromPath fail. should return 'user'")
	}
	if lib.GetUserFromUid("") != "" {
		log.Fatal("GetUserFromPath fail. should return ''")
	}
}

func TestGetFirstNElements(t *testing.T) {
	if lib.GetFirstUidElements("user", 0) != "" {
		log.Fatal("1. GetFirstPathElements fail. should return ''")
	}
	if lib.GetFirstUidElements("user", 1) != "user" {
		log.Fatal("2. GetFirstPathElements fail. should return 'user'")
	}
	if lib.GetFirstUidElements("user", 2) != "user" {
		log.Fatal("3. GetFirstPathElements fail. should return 'user'")
	}
	if lib.GetFirstUidElements("user", 3) != "user" {
		log.Fatal("4. GetFirstPathElements fail. should return 'user'")
	}
	if lib.GetFirstUidElements("user|a|b", 0) != "" {
		log.Fatal("5. GetFirstPathElements fail. should return ''")
	}
	if lib.GetFirstUidElements("user|a|b", 1) != "user" {
		log.Fatal("6. GetFirstPathElements fail. should return 'user'")
	}
	if lib.GetFirstUidElements("user|a|b", 2) != "user|a" {
		log.Fatal("7. GetFirstPathElements fail. should return 'user|a'")
	}
	if lib.GetFirstUidElements("user|a|b", 3) != "user|a|b" {
		log.Fatal("8. GetFirstPathElements fail. should return 'user|a|b'")
	}
	if lib.GetFirstUidElements("user|a|b", 4) != "user|a|b" {
		log.Fatal("9. GetUserFromPath fail. should return 'user|a|b'")
	}
	if lib.GetFirstUidElements("", 0) != "" {
		log.Fatal("10. GetFirstPathElements fail. should return ''")
	}
	if lib.GetFirstUidElements("", 1) != "" {
		log.Fatal("11. GetFirstPathElements fail. should return ''")
	}
	if lib.GetFirstUidElements("", 99) != "" {
		log.Fatal("12. GetFirstPathElements fail. should return ''")
	}
	if lib.GetFirstUidElements("user", -1) != "" {
		log.Fatal("13. GetFirstPathElements fail. should return ''")
	}
}

/*
go test -v -run TestParseJson
*/
func TestLoadAndParseJson(t *testing.T) {
	loadDataMap(dataFileName)
	testStructContains("OBJECT: N:''")
	testStructContains("OBJECT: N:'groups'")

	testStructContains("OBJECT: N:'UserA'")
	testStructContains("OBJECT: N:'UserB'")
	testStructContains("STRING: N:'post' V:'123'")
	testStructContains("STRING: N:'pre' V:'abc'")
	testStructContains("STRING: N:'pre' V:'abc'")
	testStructContains("STRING: N:'userId' V:'Buser'")
	if mapData.GetTimeStampString() != "Fri Jul 30 21:25:10 BST 2021" {
		log.Fatalf("Timestamp did not parse to UnixDate correctly. file:%s\n", dataFileName)
	}
}

func loadDataMap(fileName string) {
	if mapData == nil {
		fd := loadTestData(fileName)
		md, err := lib.NewJsonData(fd, dataMapUpdated)
		if err != nil {
			log.Fatalf("error creating new DataRoot file:%s %v\n", fileName, err)
		}
		mapData = md
		structData = parser.DiagnosticList(mapData.GetDataRoot())
	}
}

func dataMapUpdated(desc, user string, path *parser.Path, err error) {
	// fmt.Printf("Updated: %s User: %s Path:%s Err:%s\n", desc, user, path, err.Error())
}

func toJson(m parser.NodeI) string {
	return m.String()
}

func testStructContains(s string) {
	if !strings.Contains(structData, s) {
		log.Fatalf("error missing data: '%s'. file:%s\nStructure Data: %s", s, dataFileName, structData)
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
