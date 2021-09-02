package libtest

import (
	"testing"
	"time"

	"stuartdd.com/pref"
)

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
	if (tim1 % tim2) < 50 {
		t.Error("cached data should read at least 50 * faster")
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
