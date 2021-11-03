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

func (p *PrefData) AddChangeListener(cl func(string, string, string), prefix string) {
	p.changeListeners[prefix] = cl
}

func (p *PrefData) RemoveChangeListener(key string) {
	delete(p.changeListeners, key)
}

func (p *PrefData) callChangeListeners(path, value string) {
	for prefix, fn := range p.changeListeners {
		if fn != nil {
			fn(path, value, prefix)
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
	n, err := parser.CreateAndReturnNodeAtPath(p.data, parser.NewBarPath(path), parser.NT_LIST)
	if err != nil {
		return err
	}
	nL := n.(*parser.JsonList)
	var dupe parser.NodeI
	for _, v := range nL.GetValues() {
		if v.String() == value {
			dupe = v
		}
	}
	if dupe != nil {
		nL.Remove(dupe)
	}
	n.(*parser.JsonList).Add(parser.NewJsonString("", value))
	if nL.Len() > maxLen {
		nL.Remove(nL.GetNodeAt(0))
	}
	p.cache[path] = &n
	p.callChangeListeners(path, value)
	return nil
}

func (p *PrefData) PutString(path, value string) error {
	n, err := parser.CreateAndReturnNodeAtPath(p.data, parser.NewBarPath(path), parser.NT_STRING)
	if err != nil {
		return err
	}
	nS := n.(*parser.JsonString)
	if nS.GetValue() != value {
		nS.SetValue(value)
		p.cache[path] = &n
		p.callChangeListeners(path, value)
	}
	return nil
}

func (p *PrefData) PutBool(path string, value bool) error {
	n, err := parser.CreateAndReturnNodeAtPath(p.data, parser.NewBarPath(path), parser.NT_BOOL)
	if err != nil {
		return err
	}
	nB := n.(*parser.JsonBool)
	if nB.GetValue() != value {
		nB.SetValue(value)
		p.cache[path] = &n
		p.callChangeListeners(path, fmt.Sprintf("%t", value))
	}
	return nil
}

func (p *PrefData) PutFloat32(path string, value float32) error {
	return p.PutFloat64(path, float64(value))
}

func (p *PrefData) PutFloat64(path string, value float64) error {
	n, err := parser.CreateAndReturnNodeAtPath(p.data, parser.NewBarPath(path), parser.NT_NUMBER)
	if err != nil {
		return err
	}
	nF := n.(*parser.JsonNumber)
	if nF.GetValue() != value {
		nF.SetValue(value)
		p.cache[path] = &n
		p.callChangeListeners(path, fmt.Sprintf("%f", value))
	}
	return nil
}

func (p *PrefData) PutInt64(path string, value int64) error {
	n, err := parser.CreateAndReturnNodeAtPath(p.data, parser.NewBarPath(path), parser.NT_NUMBER)
	if err != nil {
		return err
	}
	nI := n.(*parser.JsonNumber)
	if nI.GetIntValue() != value {
		(n.(*parser.JsonNumber)).SetValue(float64(value))
		p.cache[path] = &n
		p.callChangeListeners(path, fmt.Sprintf("%d", value))
	}
	return nil
}

func (p *PrefData) GetBoolWithFallback(path string, fb bool) bool {
	n, ok := p.getNodeForPath(path, true)
	if ok {
		if n.GetNodeType() == parser.NT_BOOL {
			return (n.(*parser.JsonBool)).GetValue()
		}
	}
	p.PutBool(path, fb)
	return fb
}

func (p *PrefData) GetFloat64WithFallback(path string, fb float64) float64 {
	n, ok := p.getNodeForPath(path, true)
	if ok {
		if n.GetNodeType() == parser.NT_NUMBER {
			return (n.(*parser.JsonNumber)).GetValue()
		}
	}
	p.PutFloat64(path, fb)
	return fb
}

func (p *PrefData) GetInt64WithFallback(path string, fb int64) int64 {
	n, ok := p.getNodeForPath(path, true)
	if ok {
		if n.GetNodeType() == parser.NT_NUMBER {
			return (n.(*parser.JsonNumber)).GetIntValue()
		}
	}
	p.PutInt64(path, fb)
	return fb
}

func (p *PrefData) GetFloat32WithFallback(path string, fb float32) float32 {
	return float32(p.GetFloat64WithFallback(path, float64(fb)))
}

func (p *PrefData) GetFloat32WithFallbackAndMin(path string, fb, min float32) float32 {
	v := float32(p.GetFloat64WithFallback(path, float64(fb)))
	if v < min {
		return min
	}
	return v
}

func (p *PrefData) GetFloat64WithFallbackAndMin(path string, fb, min float64) float64 {
	v := p.GetFloat64WithFallback(path, fb)
	if v < min {
		return min
	}
	return v
}

func (p *PrefData) GetStringForPathWithFallback(path, fb string) string {
	n, ok := p.getNodeForPath(path, true)
	if ok {
		if n.GetNodeType() == parser.NT_STRING {
			return (n.(*parser.JsonString)).GetValue()
		}
	}
	p.PutString(path, fb)
	return fb
}

func (p *PrefData) GetDataForPath(path string) (parser.NodeI, bool) {
	return p.getNodeForPath(path, true)
}

func (p *PrefData) GetStringList(path string) []string {
	n, found := p.getNodeForPath(path, false)
	if !found {
		list := make([]string, 1)
		list[0] = ""
		return list
	}
	if n.GetNodeType() == parser.NT_OBJECT {
		return n.(*parser.JsonObject).GetSortedKeys()
	} else {
		if n.GetNodeType() == parser.NT_LIST {
			mapS := make(map[string]string)
			list := make([]string, 0)
			l := n.(*parser.JsonList).GetValues()
			for i := len(l) - 1; i >= 0; i-- {
				vS := l[i].String()
				_, ok := mapS[vS]
				if !ok {
					mapS[vS] = ""
					list = append(list, vS)
				}
			}
			return list
		}
	}
	list := make([]string, 1)
	list[0] = n.String()
	return list
}

func (p *PrefData) getNodeForPath(path string, cache bool) (parser.NodeI, bool) {
	if cache {
		cached, ok := p.cache[path]
		if ok {
			return *cached, true
		}
	}
	n, err := parser.Find(p.data, parser.NewBarPath(path))
	if err != nil || n == nil {
		return nil, false
	}
	p.cache[path] = &n
	return n, true
}
