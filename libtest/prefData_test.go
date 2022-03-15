package libtest

import (
	"fmt"
	"math"
	"os"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/stuartdd2/JsonParser4go/parser"
	"stuartdd.com/pref"
)

var (
	path1 *parser.Path = parser.NewDotPath("")
	val1  string       = ""
	path2 *parser.Path = parser.NewDotPath("")
	val2  string       = ""

	list1_fb = []string{"FB", "2", "3"}
)

func removeFile(t *testing.T, fileName string) {
	err := os.Remove(fileName)
	if err != nil {
		t.Errorf("should have removed file %s. Error: %s", fileName, err.Error())
	}
}

func TestPutStringList(t *testing.T) {
	p, _ := pref.NewPrefData("config_002.json")
	err := p.PutStringList(parser.NewDotPath("list.put1"), list1_fb, true)
	testErrorNil(t, err, "PutStringList list.put1")
	t.Errorf("%s", p.String())
	p.Save()
}

func TestGetStringList(t *testing.T) {
	p, _ := pref.NewPrefData("config_002.json")
	lx := p.GetStringListWithFallback(parser.NewDotPath("listX"), list1_fb)
	testList(t, lx, "[FB 2 3]")
	l1 := p.GetStringListWithFallback(parser.NewDotPath("list1"), list1_fb)
	testList(t, l1, "[a b c]")
	l2 := p.GetStringListWithFallback(parser.NewDotPath("list2"), list1_fb)
	testList(t, l2, "[]")
}

func testList(t *testing.T, l []string, req string) {
	ls := fmt.Sprintf("%s", l)
	if ls != req {
		t.Errorf("list '%s' should equal '%s'", l, req)
	}
}

func TestPutDropDownList(t *testing.T) {
	p, _ := pref.NewPrefData("config_001.json")
	p.AddToDropDownList(parser.NewBarPath("new|list|a"), "abc", 3)
	pp, ok := p.GetDataForPath(parser.NewBarPath("new|list|a"))
	if !ok {
		t.Error("Should find path")
	}
	if pp.(*parser.JsonList).Len() != 1 {
		t.Error("Should be len 1")
	}
	p.AddToDropDownList(parser.NewBarPath("new|list|a"), "abc", 3)
	if pp.(*parser.JsonList).Len() != 1 {
		t.Error("Should still be len 1")
	}
	p.AddToDropDownList(parser.NewBarPath("new|list|a"), "123", 3)
	if pp.(*parser.JsonList).Len() != 2 {
		t.Error("Should be len 2")
	}
	p.AddToDropDownList(parser.NewBarPath("new|list|a"), "123", 3)
	if pp.(*parser.JsonList).Len() != 2 {
		t.Error("Should still be len 2")
	}
	p.AddToDropDownList(parser.NewBarPath("new|list|a"), "123", 3)
	if pp.(*parser.JsonList).Len() != 2 {
		t.Error("Should still be len 2")
	}
	p.AddToDropDownList(parser.NewBarPath("new|list|a"), "abc", 3)
	if pp.(*parser.JsonList).Len() != 2 {
		t.Error("Should still be len 2")
	}
	p.AddToDropDownList(parser.NewBarPath("new|list|a"), "xyz1", 3)
	if pp.(*parser.JsonList).Len() != 3 {
		t.Error("Should be len 3")
	}
	p.AddToDropDownList(parser.NewBarPath("new|list|a"), "xyz2", 3)
	if pp.(*parser.JsonList).Len() != 3 {
		t.Error("Should be len 3")
	}
	l := p.GetDropDownList(parser.NewBarPath("new|list|a"))
	if fmt.Sprintf("%s", l) != "[xyz2 xyz1 abc]" {
		t.Errorf("List should be %s", l)
	}
}
func TestFloats(t *testing.T) {
	p, _ := pref.NewPrefData("TestDataTypes.json")
	p.PutFloat32(parser.NewDotPath("float.f32"), 1234.5)
	p.PutFloat64(parser.NewDotPath("float.f64"), 1234.5)

	f32 := p.GetFloat32WithFallback(parser.NewDotPath("float.f32"), float32(1.5))
	if reflect.ValueOf(f32).Kind() != reflect.Float32 {
		t.Error("should have returned float32")
	}
	if f32 != 1234.5 {
		t.Error("should have returned 1234.5")
	}
	f32 = p.GetFloat32WithFallback(parser.NewDotPath("float.f64"), float32(1.5))
	if reflect.ValueOf(f32).Kind() != reflect.Float32 {
		t.Error("should have returned float32")
	}
	if f32 != 1234.5 {
		t.Error("should have returned 1234.5")
	}

	f64 := p.GetFloat64WithFallback(parser.NewDotPath("float.f64"), float64(1.5))
	if reflect.ValueOf(f64).Kind() != reflect.Float64 {
		t.Error("should have returned float64")
	}
	if f64 != 1234.5 {
		t.Error("should have returned 1234.5")
	}
	f64 = p.GetFloat64WithFallback(parser.NewDotPath("float.f32"), float64(1.5))
	if reflect.ValueOf(f64).Kind() != reflect.Float64 {
		t.Error("should have returned float64")
	}
	if f64 != 1234.5 {
		t.Error("should have returned 1234.5")
	}
}

func TestSave(t *testing.T) {
	defer removeFile(t, "TestSaveData.txt")
	q, _ := pref.NewPrefData("TestDataTypesGold.json")
	v5 := q.GetStringWithFallback(parser.NewBarPath("groups|UserA|notes|note"), "bla")
	if v5 == "bla" {
		t.Error("should have found value")
	}
	q.SaveAs("TestSaveData.txt")

	p, _ := pref.NewPrefData("TestSaveData.txt")
	v6 := p.GetStringWithFallback(parser.NewBarPath("groups|UserA|notes|note"), "bla")
	if v6 != v5 {
		t.Error("Should be the same value")
	}

	p.PutString(parser.NewBarPath("root"), "haveatit")
	v7 := p.GetStringWithFallback(parser.NewBarPath("root"), "bla")
	if v7 != "haveatit" {
		t.Error("Should have returned haveatit")
	}

	p.Save()
	r, _ := pref.NewPrefData("TestSaveData.txt")
	v8 := r.GetStringWithFallback(parser.NewBarPath("root"), "bla")
	if v8 != "haveatit" {
		t.Error("Should have returned haveatit")
	}

}

func TestPathEquality(t *testing.T) {
	pp := parser.NewBarPath("a|b|c")
	p := parser.NewBarPath("a|b|c")
	if p.String() != "a|b|c" {
		t.Error("P String not equal")
	}
	if !p.Equal(pp) {
		t.Error("P String not equal")
	}
	if !p.Equal(parser.NewBarPath("a|b|c")) {
		t.Error("P String not equal")
	}
	if !p.Equal(parser.NewDotPath("a.b.c")) {
		t.Error("P String not equal")
	}
	if p.Equal(parser.NewBarPath("a|b")) {
		t.Error("P String not equal")
	}
}

func TestChangeListeners(t *testing.T) {
	pref, _ := pref.NewPrefData("TestDataTypes.json")
	pref.AddChangeListener(func(p *parser.Path, v string, k string) {
		path1 = p
		val1 = v
	}, "cl.")
	path1 = parser.NewDotPath("")
	val1 = ""
	path2 = parser.NewDotPath("")
	val2 = ""
	pref.PutBool(parser.NewDotPath("cl.bool.a"), true)
	if path1.String() != "cl.bool.a" || val1 != "true" {
		t.Error("Path1 and Val1 should have been updated")
	}
	if path2.String() != "" || val2 != "" {
		t.Error("Path2 and Val2 should NOT have been updated")
	}
	pref.AddChangeListener(func(p *parser.Path, v string, k string) {
		path2 = p
		val2 = v
	}, "cl.32.")
	path1 = parser.NewDotPath("")
	val1 = ""
	path2 = parser.NewDotPath("")
	val2 = ""
	pref.PutFloat32(parser.NewDotPath("cl.32.a"), 1.7)
	f1, _ := strconv.ParseFloat(val1, 32)
	f2, _ := strconv.ParseFloat(val2, 32)
	if path1.String() != "cl.32.a" || (math.Round(f1*100)/100) != 1.7 {
		t.Error("Path1 and Val1 should have been updated")
	}
	if path2.String() != "cl.32.a" || (math.Round(f2*100)/100) != 1.7 {
		t.Error("Path2 and Val2 should have been updated")
	}

}

func TestPutString(t *testing.T) {
	p, _ := pref.NewPrefData("TestDataTypesGold.json")
	err := p.PutString(parser.NewBarPath("groups|UserA|notes|note|hi"), "val")
	if err == nil {
		t.Error("should return error 'not a container node'")
	}
	v := p.GetStringWithFallback(parser.NewBarPath("groups|UserA|notes|note"), "bla")
	if v == "" || v == "bla" {
		t.Error("should have found v")
	}

	err = p.PutString(parser.NewBarPath("groups|UserA|notes|note"), "val")
	if err != nil {
		t.Error("should work")
	}

	v2 := p.GetStringWithFallback(parser.NewBarPath("groups|UserA|notes|note"), "bla")
	if v2 != "val" {
		t.Error("should have found new value")
	}

	err = p.PutString(parser.NewBarPath("groups|UserA|noes|hi"), "value3")
	if err != nil {
		t.Error("should not return an error")
	}
	v3 := p.GetStringWithFallback(parser.NewBarPath("groups|UserA|noes|hi"), "bla")
	if v3 != "value3" {
		t.Error("should have found new value (value3)")
	}

	err = p.PutString(parser.NewBarPath("groups|newUser|notes|note"), "newNote")
	if err != nil {
		t.Error("should not return an error")
	}

	v4 := p.GetStringWithFallback(parser.NewBarPath("groups|newUser|notes|note"), "bla")
	if v4 != "newNote" {
		t.Error("should have found new value (newNote)")
	}

	err = p.PutString(parser.NewBarPath("groups|newUser|notes|note"), "overwriteNote")
	if err != nil {
		t.Error("should not return an error")
	}

	v5 := p.GetStringWithFallback(parser.NewBarPath("groups|newUser|notes|note"), "bla")
	if v5 != "overwriteNote" {
		t.Error("should have found new value (overwriteNote)")
	}

	err = p.PutString(parser.NewBarPath("newRoot"), "newRootValue")
	if err != nil {
		t.Error("should not return an error")
	}

	v6 := p.GetStringWithFallback(parser.NewBarPath("newRoot"), "bla")
	if v6 != "newRootValue" {
		t.Error("should have found new value (newRootValue)")
	}

	err = p.PutString(parser.NewBarPath(".xRoot"), "dotRootValue")
	if err != nil {
		t.Error("should not return an error")
	}

	v7 := p.GetStringWithFallback(parser.NewBarPath(".xRoot"), "bla")
	if v7 != "dotRootValue" {
		t.Error("should have found new value (dotRootValue)")
	}
}
func TestLoadFallback(t *testing.T) {
	p, err := pref.NewPrefData("TestDataTypesGold.json")
	if err != nil {
		t.Error("should NOT return error")
	}
	if p.GetFileName() != "TestDataTypesGold.json" {
		t.Error("file name was not stored correctly")
	}
	m, _ := p.GetDataForPath(parser.NewDotPath("groups.UserA.notes.note"))
	if m.String() == "" {
		t.Error("groups|UserA|notes|note should return a value")
	}
	s2 := p.GetStringWithFallback(parser.NewBarPath("groups|UserA|notes|note"), "x")
	if m.String() != s2 {
		t.Error("GetStringWithFallback should return same as GetDataForPath")
	}
	s3 := p.GetStringWithFallback(parser.NewBarPath("groups|UserA|notes|not"), "fallback")
	if s3 != "fallback" {
		t.Error("groups|UserA|notes|not should return \"fallback\" ")
	}

}
func TestLoadCacheAfterPut(t *testing.T) {
	p, err := pref.NewPrefData("TestDataTypes.json")
	if err != nil {
		t.Error("should NOT return error")
	}
	if p.GetFileName() != "TestDataTypes.json" {
		t.Error("file name was not stored correctly")
	}
	p.PutString(parser.NewBarPath("a|b|c"), "abc")
	s := p.GetStringWithFallback(parser.NewBarPath("a|b|c"), "xyz")
	if s != "abc" {
		t.Error("Incorrect value returned. Not abc")
	}
	p.PutString(parser.NewBarPath("a|b|c"), "123")
	s = p.GetStringWithFallback(parser.NewBarPath("a|b|c"), "xyz")
	if s != "123" {
		t.Error("Incorrect value returned. Not 123")
	}
}

func TestLoadCache(t *testing.T) {
	p, err := pref.NewPrefData("TestDataTypesGold.json")
	if err != nil {
		t.Error("should NOT return error")
	}
	if p.GetFileName() != "TestDataTypesGold.json" {
		t.Error("file name was not stored correctly")
	}
	sta1 := time.Now().UnixNano()
	m1, ok1 := p.GetDataForPath(parser.NewBarPath("groups|UserA|notes|note"))
	timUnCached := time.Now().UnixNano() - sta1

	for i := 0; i < 5; i++ {
		p.GetDataForPath(parser.NewBarPath("groups|UserA|notes|note"))
	}

	sta2 := time.Now().UnixNano()
	m2, ok2 := p.GetDataForPath(parser.NewBarPath("groups|UserA|notes|note"))
	timCached := time.Now().UnixNano() - sta2

	if m1.String() == "" {
		t.Error("v1 groups|UserA|notes|note should return a string")
	}
	if m2.String() == "" {
		t.Error("v2 groups|UserA|notes|note should return a string")
	}
	if !ok1 {
		t.Error("ok1 groups|UserA|notes|note should return true")
	}
	if !ok2 {
		t.Error("ok2 groups|UserA|notes|note should return true")
	}

	if m1 != m2 {
		t.Error("cached data should return the same value")
	}
	diff := timUnCached / timCached
	if (diff) < 5 {
		t.Errorf("cached data should read at least 5 * faster. Actual: %d", diff)
	}
}

func TestLoadComplex(t *testing.T) {
	p, err := pref.NewPrefData("TestDataTypesGold.json")
	if err != nil {
		t.Error("should NOT return error")
	}
	if p.GetFileName() != "TestDataTypesGold.json" {
		t.Error("file name was not stored correctly")
	}

	m, ok := p.GetDataForPath(parser.NewDotPath("groups.fred"))
	if m != nil {
		t.Error("groups.fred should return empty string")
	}
	if ok {
		t.Error("ok groups.fred should return false")
	}
	if m != nil {
		t.Error("groups should not return a map")
	}

	m, ok = p.GetDataForPath(parser.NewBarPath("groups|UserA|notes|fred"))
	if m != nil {
		t.Error("groups|UserA|notes|fred should return empty string")
	}
	if ok {
		t.Error("ok groups|UserA|notes|fred should return false")
	}
	if m != nil {
		t.Error("groups|UserA|notes|fred should not return a map")
	}

	m, ok = p.GetDataForPath(parser.NewBarPath("groups"))
	if m.GetNodeType() != parser.NT_OBJECT {
		t.Error("groups should be a JsonObject")
	}
	if !ok {
		t.Error("ok groups should return true")
	}
	if m == nil {
		t.Error("groups should return a map")
	}

	m, ok = p.GetDataForPath(parser.NewBarPath("groups|UserA|notes"))
	if !ok {
		t.Error("ok groups|UserA|notes should return true")
	}
	if m.GetNodeType() != parser.NT_OBJECT {
		t.Error("groups should  be a JsonObject")
	}
	if m == nil {
		t.Error("groups should return a map")
	}
	m, ok = p.GetDataForPath(parser.NewBarPath("groups|UserA|notes|note"))
	if !ok {
		t.Error("ok groups|UserA|notes|not should return true")
	}
	if m.String() != "An amazing A note (dont panic) fdf" {
		t.Error("groups|UserA|notes|note should return  'An amazing A note (dont panic) fdf'")
	}
	if m == nil {
		t.Error("groups should return a node")
	}
}

func TestLoadSinglePath(t *testing.T) {
	p, err := pref.NewPrefData("config_001.json")
	if err != nil {
		t.Error("should NOT return error")
	}
	if p.GetFileName() != "config_001.json" {
		t.Error("file name was not stored correctly")
	}
	m, ok := p.GetDataForPath(parser.NewBarPath("boolean"))
	if !ok {
		t.Error("ok boolean should return true")
	}
	if m.String() != "true" {
		t.Error("boolean should return 'true'")
	}
	if m == nil {
		t.Error("boolean should return node")
	}
	m, ok = p.GetDataForPath(parser.NewBarPath("split"))
	if !ok {
		t.Error("ok split should return true")
	}
	if m.String() != "0.2" {
		t.Error("split should return '0.2'")
	}
	if m == nil {
		t.Error("split should return node")
	}
	m, ok = p.GetDataForPath(parser.NewBarPath("integer"))
	if !ok {
		t.Error("ok integer should return true")
	}
	if m.String() != "830" {
		t.Error("integer should return '830'")
	}
	if m == nil {
		t.Error("integer should return node")
	}
	m, ok = p.GetDataForPath(parser.NewBarPath("float"))
	if !ok {
		t.Error("ok float should return true")
	}
	if m.String() != "479.52" {
		t.Error("float should return '479.52'")
	}
	if m == nil {
		t.Error("float should return node")
	}
}

func TestLoadError(t *testing.T) {
	_, err := pref.NewPrefData("ABC")
	if err == nil {
		t.Error("should return found")
	}
	_, err = pref.NewPrefData("config_notjson.json")
	if err == nil {
		t.Error("should return found")
	}
}
