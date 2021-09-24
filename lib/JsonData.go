package lib

import (
	"fmt"
	"sort"
	"time"

	"github.com/stuartdd/jsonParserGo/parser"
)

type JsonData struct {
	timeStamp      time.Time
	dataMap        parser.NodeI
	navIndex       map[string][]string
	dataMapUpdated func(string, string, string, error)
}

func NewJsonData(j []byte, dataMapUpdated func(string, string, string, error)) (*JsonData, error) {
	m, err := parser.Parse(j)
	if err != nil {
		return nil, err
	}
	ts, err := parser.Find(m, "timeStamp")
	if err != nil {
		return nil, fmt.Errorf("'%s' does not exist in data root", timeStampStr)
	}
	tim, err := parseTime((ts.(*parser.JsonString)).GetValue())
	if err != nil {
		return nil, fmt.Errorf("'%s' could not be parsed", timeStampStr)
	}
	// _, ok = m[dataMapRootName]
	// if !ok {
	// 	return nil, fmt.Errorf("'%s' does not exist in data root", dataMapRootName)
	// }

	dr := &JsonData{timeStamp: tim, dataMap: m, navIndex: *createNavIndex2(m), dataMapUpdated: dataMapUpdated}
	return dr, nil
}

func (r *JsonData) GetTimeStamp2() time.Time {
	return r.timeStamp
}

func (r *JsonData) GetNavIndex2(id string) []string {
	return r.navIndex[id]
}

func createNavIndex2(m parser.NodeI) *map[string][]string {
	var uids = make(map[string][]string)
	createNavIndexDetail2("", &uids, m)
	return &uids
}

func createNavIndexDetail2(id string, uids *map[string][]string, nodeI parser.NodeI) {
	if nodeI.GetNodeType() == parser.NT_LIST {
		panic("createNavIndexDetail2: cannot process lists")
	}
	g, err := parser.Find(nodeI, "groups")
	if err != nil {
		panic("createNavIndexDetail2: cannot find groups")
	}
	userObj := g.(*parser.JsonObject)
	(*uids)[""] = userObj.GetSortedKeys()
	users := userObj.GetValuesSorted()
	for _, user := range users {
		if user.GetNodeType() == parser.NT_OBJECT {
			userO := user.(*parser.JsonObject)
			id := userO.GetName()
			l, ll := keysToList2(id, userO)
			(*uids)[id] = ll
			for ii2, id2 := range l {
				m2 := userO.GetNodeWithName(id2).(*parser.JsonObject)
				for _, v2 := range m2.GetValuesSorted() {
					if v2.GetNodeType() == parser.NT_OBJECT {
						l2, ll2 := keysToList2(ll[ii2], m2)
						(*uids)[ll[ii2]] = ll2
						for ii3, id3 := range l2 {
							m3 := m2.GetNodeWithName(id3).(*parser.JsonObject)
							for _, v3 := range m3.GetValuesSorted() {
								if v3.GetNodeType() == parser.NT_OBJECT {
									_, ll3 := keysToList2(ll2[ii3], m3)
									(*uids)[ll2[ii3]] = ll3
								}
							}
						}
					}
				}
			}
		}
	}
}

func keysToList2(id string, m *parser.JsonObject) ([]string, []string) {
	l := make([]string, 0)
	ll := make([]string, 0)
	for _, k := range m.GetSortedKeys() {
		l = append(l, k)
		if id == "" {
			ll = append(ll, k)
		} else {
			ll = append(ll, fmt.Sprintf("%s.%s", id, k))
		}
	}
	sort.Strings(l)
	sort.Strings(ll)
	return l, ll
}
