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

func (p *PrefData) Save() error {
	return p.SaveAs(p.fileName)
}

func (p *PrefData) SaveAs(fileName string) error {
	return ioutil.WriteFile(fileName, []byte(p.String()), 0644)
}

func (p *PrefData) GetFileName() string {
	return p.fileName
}

func (p *PrefData) String() string {
	output, err := json.MarshalIndent(p.data, "", "    ")
	if err != nil {
		return ""
	}
	return string(output)
}

func (p *PrefData) PutRootString(name, value string) error {
	return p.PutString("", name, value)
}

func (p *PrefData) PutString(path, name, value string) error {
	m, s, ok := p.getDataForPath(path, false)
	if ok && s != "" {
		return fmt.Errorf("path %s is an end (leaf) node already", path)
	}
	if m == nil {
		p.makePath(path)
		m, _, _ = p.getDataForPath(path, false)
		if m == nil {
			return fmt.Errorf("Failed to create node in path %s", path)
		}
	}
	(*m)[name] = value
	p.cache[path+"."+name] = value
	return nil
}

func (p *PrefData) GetBoolWithFallback(name string, fb bool) bool {
	s := p.GetValueForPathWithFallback(name, fmt.Sprintf("%t", fb))
	s = strings.ToLower(s)
	if strings.HasPrefix(s, "tr") {
		return true
	}
	return false
}

func (p *PrefData) GetValueForPathWithFallback(path, fb string) string {
	_, v, ok := p.getDataForPath(path, true)
	if ok {
		return v
	}
	return fb
}

func (p *PrefData) GetDataForPath(path string) (*map[string]interface{}, string, bool) {
	return p.getDataForPath(path, true)
}

func (p *PrefData) makePath(path string) error {
	nodes := strings.Split(path, ".")
	x := p.data
	for _, v := range nodes {
		y := x[v]
		if y == nil {
			x[v] = make(map[string]interface{})
			y = x[v]
			x = y.(map[string]interface{})
		} else {
			x = y.(map[string]interface{})
		}
	}
	return nil
}

func (p *PrefData) getDataForPath(path string, cache bool) (*map[string]interface{}, string, bool) {
	if cache {
		cached, ok := p.cache[path]
		if ok {
			return nil, cached, true
		}
	}
	nodes := strings.Split(path, ".")
	if nodes[0] == "" {
		return &p.data, "", true
	}
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

func getParentPath(path string) string {
	if path == "" {
		return path
	}
	p := strings.LastIndexByte(path, '.')
	switch p {
	case -1:
		return ""
	case 0:
		return ""
	default:
		return path[0:p]
	}
}
