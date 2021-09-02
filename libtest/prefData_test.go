package libtest

import (
	"os"
	"testing"
	"time"

	"stuartdd.com/pref"
)

func removeFile(t *testing.T, fileName string) {
	err := os.Remove(fileName)
	if err != nil {
		t.Errorf("should have removed file %s. Error: %s", fileName, err.Error())
	}
}

func TestSave(t *testing.T) {
	defer removeFile(t, "TestSaveData.txt")
	q, _ := pref.NewPrefData("TestDataTypes.json")
	v5 := q.GetValueForPathWithFallback("groups.UserA.notes.note", "bla")
	if v5 == "bla" {
		t.Error("should have found value")
	}
	q.SaveAs("TestSaveData.txt")

	p, _ := pref.NewPrefData("TestSaveData.txt")
	v6 := p.GetValueForPathWithFallback("groups.UserA.notes.note", "bla")
	if v6 != v5 {
		t.Error("Should be the same value")
	}

	p.PutString("", "root", "haveatit")
	v7 := p.GetValueForPathWithFallback("root", "bla")
	if v7 != "haveatit" {
		t.Error("Should have returned haveatit")
	}

	p.Save()
	r, _ := pref.NewPrefData("TestSaveData.txt")
	v8 := r.GetValueForPathWithFallback("root", "bla")
	if v8 != "haveatit" {
		t.Error("Should have returned haveatit")
	}

}
func TestPutString(t *testing.T) {
	p, _ := pref.NewPrefData("TestDataTypes.json")
	err := p.PutString("groups.UserA.notes.note", "hi", "val")
	if err == nil {
		t.Error("should return error 'Path x is an end")
	}
	v := p.GetValueForPathWithFallback("groups.UserA.notes.note", "bla")
	if v == "" || v == "bla" {
		t.Error("should have found v")
	}

	err = p.PutString("groups.UserA.notes", "note", "val")
	if err != nil {
		t.Error("should work")
	}

	v2 := p.GetValueForPathWithFallback("groups.UserA.notes.note", "bla")
	if v2 != "val" {
		t.Error("should have found new value")
	}

	err = p.PutString("groups.UserA.noes", "hi", "value3")
	if err != nil {
		t.Error("should not return an error")
	}
	v3 := p.GetValueForPathWithFallback("groups.UserA.noes.hi", "bla")
	if v3 != "value3" {
		t.Error("should have found new value (value3)")
	}

	err = p.PutString("groups.newUser.notes", "note", "newNote")
	if err != nil {
		t.Error("should not return an error")
	}

	v4 := p.GetValueForPathWithFallback("groups.newUser.notes.note", "bla")
	if v4 != "newNote" {
		t.Error("should have found new value (newNote)")
	}

	err = p.PutString("groups.newUser.notes", "note", "overwriteNote")
	if err != nil {
		t.Error("should not return an error")
	}

	v5 := p.GetValueForPathWithFallback("groups.newUser.notes.note", "bla")
	if v5 != "overwriteNote" {
		t.Error("should have found new value (overwriteNote)")
	}

	err = p.PutString("", "newRoot", "newRootValue")
	if err != nil {
		t.Error("should not return an error")
	}

	v6 := p.GetValueForPathWithFallback("newRoot", "bla")
	if v6 != "newRootValue" {
		t.Error("should have found new value (newRootValue)")
	}

}
func TestLoadFallback(t *testing.T) {
	p, err := pref.NewPrefData("TestDataTypes.json")
	if err != nil {
		t.Error("should NOT return error")
	}
	if p.GetFileName() != "TestDataTypes.json" {
		t.Error("file name was not stored correctly")
	}
	_, s1, _ := p.GetDataForPath("groups.UserA.notes.note")
	if s1 == "" {
		t.Error("groups.UserA.notes.note should return a value")
	}
	s2 := p.GetValueForPathWithFallback("groups.UserA.notes.note", "x")
	if s1 != s2 {
		t.Error("GetValueForPathWithFallback should return same as GetDataForPath")
	}
	s3 := p.GetValueForPathWithFallback("groups.UserA.notes.not", "fallback")
	if s3 != "fallback" {
		t.Error("groups.UserA.notes.not should return \"fallback\" ")
	}

}
func TestLoadCache(t *testing.T) {
	p, err := pref.NewPrefData("TestDataTypes.json")
	if err != nil {
		t.Error("should NOT return error")
	}
	if p.GetFileName() != "TestDataTypes.json" {
		t.Error("file name was not stored correctly")
	}
	sta1 := time.Now().UnixNano()
	_, v1, ok1 := p.GetDataForPath("groups.UserA.notes.note")
	tim1 := time.Now().UnixNano() - sta1
	sta2 := time.Now().UnixNano()
	_, v2, ok2 := p.GetDataForPath("groups.UserA.notes.note")
	tim2 := time.Now().UnixNano() - sta2

	if v1 == "" {
		t.Error("v1 groups.UserA.notes.note should return a string")
	}
	if v2 == "" {
		t.Error("v2 groups.UserA.notes.note should return a string")
	}
	if !ok1 {
		t.Error("ok1 groups.UserA.notes.note should return true")
	}
	if !ok2 {
		t.Error("ok2 groups.UserA.notes.note should return true")
	}

	if v2 != v1 {
		t.Error("cached data should return the same value")
	}
	if (tim1 % tim2) < 40 {
		t.Error("cached data should read at least 40 * faster")
	}
}

func TestLoadComplex(t *testing.T) {
	p, err := pref.NewPrefData("TestDataTypes.json")
	if err != nil {
		t.Error("should NOT return error")
	}
	if p.GetFileName() != "TestDataTypes.json" {
		t.Error("file name was not stored correctly")
	}

	m, s, ok := p.GetDataForPath("groups.fred")
	if s != "" {
		t.Error("groups.fred should return empty string")
	}
	if ok {
		t.Error("ok groups.fred should return false")
	}
	if m != nil {
		t.Error("groups should not return a map")
	}

	m, s, ok = p.GetDataForPath("groups.UserA.notes.fred")
	if s != "" {
		t.Error("groups.UserA.notes.fred should return empty string")
	}
	if ok {
		t.Error("ok groups.UserA.notes.fred should return false")
	}
	if m != nil {
		t.Error("groups.UserA.notes.fred should not return a map")
	}

	m, s, ok = p.GetDataForPath("groups")
	if s != "" {
		t.Error("groups should return empty string")
	}
	if !ok {
		t.Error("ok groups should return true")
	}
	if m == nil {
		t.Error("groups should return a map")
	}
	if len(*m) != 2 {
		t.Error("groups map is len 2")
	}

	m, s, ok = p.GetDataForPath("groups.UserA.notes")
	if !ok {
		t.Error("ok groups.UserA.notes should return true")
	}
	if s != "" {
		t.Error("groups should return empty string")
	}
	if m == nil {
		t.Error("groups should return a map")
	}
	if len(*m) != 2 {
		t.Error("groups map is len 2")
	}
	note := (*m)["note"]
	if note != "An amazing A note (dont panic) fdf" {
		t.Error("groups map is len 2")
	}
	m, s, ok = p.GetDataForPath("groups.UserA.notes.note")
	if !ok {
		t.Error("ok groups.UserA.notes.not should return true")
	}
	if s != "An amazing A note (dont panic) fdf" {
		t.Error("groups.UserA.notes.note should return  'An amazing A note (dont panic) fdf'")
	}
	if m != nil {
		t.Error("groups should return a map")
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
	m, s, ok := p.GetDataForPath("boolean")
	if !ok {
		t.Error("ok boolean should return true")
	}
	if s != "true" {
		t.Error("boolean should return 'true'")
	}
	if m != nil {
		t.Error("boolean should return nil map")
	}
	m, s, ok = p.GetDataForPath("split")
	if !ok {
		t.Error("ok split should return true")
	}
	if s != "0.2" {
		t.Error("split should return '0.2'")
	}
	if m != nil {
		t.Error("split should return nil map")
	}
	m, s, ok = p.GetDataForPath("integer")
	if !ok {
		t.Error("ok integer should return true")
	}
	if s != "830" {
		t.Error("integer should return '830'")
	}
	if m != nil {
		t.Error("integer should return nil map")
	}
	m, s, ok = p.GetDataForPath("float")
	if !ok {
		t.Error("ok float should return true")
	}
	if s != "479.52" {
		t.Error("float should return '479.52'")
	}
	if m != nil {
		t.Error("float should return nil map")
	}
	ss := p.String()
	if ss != "{\"boolean\":true,\"float\":479.52,\"integer\":830,\"split\":0.2}" {
		t.Errorf("String returned %s NOT {\"boolean\":true,\"float\":479.52,\"integer\":830,\"split\":0.2}", ss)
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
