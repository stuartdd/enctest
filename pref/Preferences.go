package pref

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"io/ioutil"
)

type PrefData struct {
	fileName        string
	data            map[string]interface{}
	cache           map[string]string
	changeListeners map[string]func(string, string, string)
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
	cl := make(map[string]func(string, string, string), 0)
	return &PrefData{fileName: fileName, data: m, cache: c, changeListeners: cl}, nil
}

func (p *PrefData) AddChangeListener(cl func(string, string, string), filter string) {
	p.changeListeners[filter] = cl
}

func (p *PrefData) RemoveChangeListener(key string) {
	delete(p.changeListeners, key)
}

func (p *PrefData) callChangeListeners(path, value string) {
	for filter, fn := range p.changeListeners {
		if fn != nil {
			if strings.HasPrefix(path, filter) {
				fn(path, value, filter)
			}
		}
	}
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

func (p *PrefData) PutString(path, value string) error {
	parent, name := getParentAndName(path)
	m, s, ok := p.getDataForPath(parent, false)
	if ok && s != "" {
		return fmt.Errorf("path %s is an end (leaf) node already", parent)
	}
	if m == nil {
		p.makePath(parent)
		m, _, _ = p.getDataForPath(parent, false)
		if m == nil {
			return fmt.Errorf("failed to create node for path %s", path)
		}
	}
	(*m)[name] = value
	p.cache[path] = value
	p.callChangeListeners(path, value)
	return nil
}

func (p *PrefData) PutBool(path string, value bool) error {
	return p.PutString(path, fmt.Sprintf("%t", value))
}

func (p *PrefData) PutFloat32(path string, value float32) error {
	return p.PutString(path, fmt.Sprintf("%f", value))
}

func (p *PrefData) PutFloat64(path string, value float64) error {
	return p.PutString(path, fmt.Sprintf("%f", value))
}

func (p *PrefData) GetBoolWithFallback(path string, fb bool) bool {
	s := strings.ToLower(p.GetStringForPathWithFallback(path, fmt.Sprintf("%t", fb)))
	return strings.HasPrefix(s, "tr")
}

func (p *PrefData) GetFloat64WithFallback(path string, fb float64) float64 {
	s := p.GetStringForPathWithFallback(path, fmt.Sprintf("%f", fb))
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return fb
	}
	return f
}

func (p *PrefData) GetFloat32WithFallback(path string, fb float32) float32 {
	s := p.GetStringForPathWithFallback(path, fmt.Sprintf("%f", fb))
	f, err := strconv.ParseFloat(s, 32)
	if err != nil {
		return fb
	}
	return float32(f)
}

func (p *PrefData) GetStringForPathWithFallback(path, fb string) string {
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

func getParentAndName(path string) (string, string) {
	if path == "" {
		return "", ""
	}
	p := strings.LastIndexByte(path, '.')
	switch p {
	case -1:
		return "", path
	default:
		return path[0:p], path[p+1:]
	}
}
