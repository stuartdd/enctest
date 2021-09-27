package lib

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/stuartdd/jsonParserGo/parser"
)

const (
	hintStr         = "pwHints"
	noteStr         = "notes"
	dataMapRootName = "groups"
	timeStampStr    = "timeStamp"
	tabdata         = "                                     "
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
	ts, err := parser.Find(m, timeStampStr)
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

	dr := &JsonData{timeStamp: tim, dataMap: m, navIndex: *createNavIndex(m), dataMapUpdated: dataMapUpdated}
	return dr, nil
}

func (r *JsonData) GetTimeStamp() time.Time {
	return r.timeStamp
}

func (r *JsonData) GetNavIndex(id string) []string {
	return r.navIndex[id]
}

func (r *JsonData) GetDataRoot() parser.NodeI {
	return r.dataMap
}

func (r *JsonData) ToJson() string {
	return r.dataMap.String()
}

func (p *JsonData) Search(addPath func(string, string), needle string, matchCase bool) {
	groups := p.dataMap.(*parser.JsonObject).GetNodeWithName(dataMapRootName).(*parser.JsonObject)
	for _, v := range groups.GetValues() {
		searchUsers(addPath, needle, v.GetName(), v.(*parser.JsonObject), matchCase)
	}
}

func searchUsers(addPath func(string, string), needle, user string, m *parser.JsonObject, matchCase bool) {
	for _, v := range m.GetValues() {
		if v.GetName() == hintStr {
			for _, v1 := range v.(*parser.JsonObject).GetValues() {
				searchLeafNodes(addPath, true, needle, user, v1.GetName(), v1.(*parser.JsonObject), matchCase)
			}
		} else {
			searchLeafNodes(addPath, false, needle, user, v.GetName(), v.(*parser.JsonObject), matchCase)
		}
	}
}

func searchLeafNodes(addPath func(string, string), isHint bool, needle, user, name string, m *parser.JsonObject, matchCase bool) {
	tag1 := "."
	if isHint {
		tag1 = "." + hintStr + "."
	}
	if containsWithCase(name, needle, matchCase) {
		addPath(user+tag1+name, searchDeriveText(user, isHint, name, "LHS Tree", ""))
	}
	for _, s := range m.GetValues() {
		if containsWithCase(s.GetName(), needle, matchCase) {
			addPath(user+tag1+name, searchDeriveText(user, isHint, name, "Field Name", s.GetName()))
		}
		if s.GetNodeType() == parser.NT_STRING {
			if containsWithCase(s.(*parser.JsonString).GetValue(), needle, matchCase) {
				addPath(user+tag1+name, searchDeriveText(user, isHint, name, "In Text", s.GetName()))
			}
		} else {
			searchLeafNodes(addPath, isHint, needle, user, s.GetName(), s.(*parser.JsonObject), matchCase)
		}
	}
}

func searchDeriveText(user string, isHint bool, name, desc, key string) string {
	if key != "" {
		key = "'" + key + "'"
	}
	if isHint {
		return user + " [Hint] " + name + ":  " + desc + ": " + key
	}
	if name == "notes" {
		return user + " [Notes] :  " + desc + ": " + key
	}
	return user + " " + name + ":  " + desc + ": " + key
}

func containsWithCase(haystack, needle string, matchCase bool) bool {
	if matchCase {
		return strings.Contains(haystack, needle)
	} else {
		h := strings.ToLower(haystack)
		n := strings.ToLower(needle)
		return strings.Contains(h, n)
	}
}

func GetMapForUid(uid string, m parser.NodeI) (parser.NodeI, error) {
	nodes, err := parser.Find(m, uid)
	if err != nil {
		return nil, err
	}
	return nodes, nil
}

func createNavIndex(m parser.NodeI) *map[string][]string {
	var uids = make(map[string][]string)
	createNavIndexDetail("", &uids, m)
	return &uids
}

func createNavIndexDetail(id string, uids *map[string][]string, nodeI parser.NodeI) {
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

func parseTime(st string) (time.Time, error) {
	t, err := time.Parse(time.UnixDate, st)
	if err != nil {
		return time.Now(), err
	}
	return t, nil
}

func GetLastId(uid string) string {
	if uid == "" {
		return uid
	}
	p := strings.LastIndexByte(uid, '.')
	switch p {
	case -1:
		return uid
	case 0:
		return uid[1:]
	default:
		return uid[p+1:]
	}
}

func GetParentId(uid string) string {
	if uid == "" {
		return uid
	}
	p := strings.LastIndexByte(uid, '.')
	switch p {
	case -1:
		return "groups"
	case 0:
		return ""
	default:
		return uid[0:p]
	}
}

func GetPathElementAt(path string, index int) string {
	elements := strings.Split(path, ".")
	l := len(elements)
	if l == 0 || index < 0 || index >= l {
		return ""
	}
	if index < l {
		return elements[index]
	}
	return elements[l-1]
}

func GetUserFromPath(path string) string {
	return GetFirstPathElements(path, 1)
}

func GetHintFromPath(path string) string {
	return GetPathElementAt(path, 2)
}

func GetFirstPathElements(path string, count int) string {
	if count <= 0 {
		return ""
	}
	var sb strings.Builder
	dotCount := 0
	for _, c := range path {
		if c == '.' {
			dotCount++
		}
		if dotCount == count {
			return sb.String()
		}
		sb.WriteByte(byte(c))
	}
	return sb.String()
}

func (p *JsonData) GetRootUidOrCurrentUid(currentUid string) string {
	if currentUid != "" {
		for i := 4; i > 0; i-- {
			x := GetFirstPathElements(currentUid, i)
			_, ok := p.navIndex[x]
			if ok {
				return currentUid
			}
		}
	}
	l := make([]string, 0)
	for k := range p.navIndex {
		if k != "" {
			l = append(l, k)
		}
	}
	if len(l) > 0 {
		sort.Strings(l)
		return l[0]
	}
	return ""
}
