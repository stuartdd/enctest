/*
 * Copyright (C) 2021 Stuart Davies (stuartdd)
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */
package pref

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/stuartdd/jsonParserGo/parser"
)

type PrefData struct {
	fileName        string
	data            *parser.JsonObject
	cache           map[string]*parser.NodeI
	changeListeners map[string]func(string, string, string)
}

func NewPrefData(fileName string) (*PrefData, error) {
	j, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	data, err := parser.Parse(j)
	if err != nil {
		return nil, err
	}
	if data.GetNodeType() != parser.NT_OBJECT {
		return nil, fmt.Errorf("error reading '%s'. Config data root node MUST be a JsonObject type", fileName)
	}
	c := make(map[string]*parser.NodeI)
	cl := make(map[string]func(string, string, string))
	return &PrefData{fileName: fileName, data: data.(*parser.JsonObject), cache: c, changeListeners: cl}, nil
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
	return ioutil.WriteFile(fileName, []byte(p.data.JsonValueIndented(4)), 0644)
}

func (p *PrefData) GetFileName() string {
	return p.fileName
}

func (p *PrefData) String() string {
	return string(p.data.JsonValueIndented(4))
}

func (p *PrefData) PutStringList(path, value string, maxLen int) error {
	m, _, err := p.createAndReturnNodeAtPath(path, parser.NT_LIST)
	if err != nil {
		return err
	}
	mL := m.(*parser.JsonList)
	mL.Add(parser.NewJsonString("", value))
	return nil
}

func (p *PrefData) PutString(path, value string) error {
	m, _, err := p.createAndReturnNodeAtPath(path, parser.NT_STRING)
	if err != nil {
		return err
	}
	m.(*parser.JsonString).SetValue(value)
	p.cache[path] = &m
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
	_, v, ok := p.getNodeForPath(path, true)
	if ok {
		return v
	}
	return fb
}

func (p *PrefData) GetDataForPath(path string) (parser.NodeI, string, bool) {
	return p.getNodeForPath(path, true)
}

func (p *PrefData) GetStringList(path string) []string {
	n, _, found := p.getNodeForPath(path, false)
	if !found {
		list := make([]string, 1)
		list[0] = ""
		return list
	}
	if n.GetNodeType() == parser.NT_OBJECT {
		return n.(*parser.JsonObject).GetSortedKeys()
	}
	if n.GetNodeType() == parser.NT_LIST {
		list := make([]string, 0)
		l := n.(*parser.JsonList).GetValues()
		for _, v := range l {
			list = append(list, v.String())
		}
		return list
	}
	list := make([]string, 1)
	list[0] = n.String()
	return list
}

func (p *PrefData) createAndReturnNodeAtPath(path string, nodeType parser.NodeType) (parser.NodeI, string, error) {
	if path == "" {
		return nil, "", fmt.Errorf("cannot create a node from an empty path")
	}
	rootPath, name := getParentAndNameFromPath(path)
	if rootPath == "" {
		ret := parser.NewJsonType(name, nodeType)
		p.data.Add(ret)
		return ret, name, nil
	}
	cNode := p.data
	paths := strings.Split(rootPath, ".")

	for _, nn := range paths {
		n := cNode.GetNodeWithName(nn)
		if n == nil {
			n = parser.NewJsonObject(nn)
			cNode.Add(n)
		}
		if n.GetNodeType() != parser.NT_OBJECT {
			return nil, "", fmt.Errorf("found node at [%s] but it is not a container node", nn)
		}
		cNode = n.(*parser.JsonObject)
	}
	ret := cNode.GetNodeWithName(name)
	if ret == nil {
		ret = parser.NewJsonType(name, nodeType)
		cNode.Add(ret)
	}
	return ret, name, nil
}

func (p *PrefData) getNodeForPath(path string, cache bool) (parser.NodeI, string, bool) {
	if cache {
		cached, ok := p.cache[path]
		if ok {
			return *cached, (*cached).String(), true
		}
	}
	n, err := parser.Find(p.data, path)
	if err != nil {
		return nil, "", false
	}
	p.cache[path] = &n
	if n.GetNodeType() == parser.NT_LIST || n.GetNodeType() == parser.NT_OBJECT {
		return n, "", true
	}
	return n, n.String(), true
}

func getParentAndNameFromPath(path string) (string, string) {
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
