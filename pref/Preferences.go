package pref

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"io/ioutil"
)

type PrefData struct {
	fileName string
	data     map[string]interface{}
	cache    map[string]string
}

func NewPrefData(fileName string) (*PrefData, error) {
	j, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	var m map[string]interface{}
	err = json.Unmarshal([]byte(j), &m)
	if err != nil {
		return nil, err
	}
	c := make(map[string]string)
	return &PrefData{fileName: fileName, data: m, cache: c}, nil
}

func (p *PrefData) GetValueForPathWithFallback(path, fb string) string {
	_, v, ok := p.GetDataForPath(path)
	if ok {
		return v
	}
	return fb
}

func (p *PrefData) GetDataForPath(path string) (*map[string]interface{}, string, bool) {
	cached, ok := p.cache[path]
	if ok {
		return nil, cached, true
	}
	nodes := strings.Split(path, ".")
	x := p.data
	for _, v := range nodes {
		y := x[v]
		if y == nil {
			return nil, "", false
		}
		if reflect.TypeOf(y).Kind() == reflect.Map {
			x = y.(map[string]interface{})
		} else {
			val := fmt.Sprintf("%v", y)
			p.cache[path] = val
			return nil, val, true
		}
	}
	if x == nil {
		return nil, "", false
	}
	return &x, "", true
}

func (p *PrefData) GetFileName() string {
	return p.fileName
}

func (p *PrefData) String() string {
	output, err := json.Marshal(p.data)
	if err != nil {
		return ""
	}
	return string(output)
}
