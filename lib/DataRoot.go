package lib

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"
)

const (
	hintStr      = "pwHints"
	noteStr      = "notes"
	groupsStr    = "groups"
	timeStampStr = "timeStamp"
)

var (
	tabdata = "                                     "
)

type DataRoot struct {
	timeStamp      time.Time
	dataMap        map[string]interface{}
	navIndex       map[string][]string
	dataMapUpdated func(string, string, string, error)
}

func NewDataRoot(j []byte, dataMapUpdated func(string, string, string, error)) (*DataRoot, error) {
	m, err := parseJson(j)
	if err != nil {
		return nil, err
	}
	ts, ok := m[timeStampStr]
	if !ok {
		return nil, fmt.Errorf("'%s' does not exist in data root", timeStampStr)
	}
	if reflect.ValueOf(ts).Kind() != reflect.String {
		return nil, fmt.Errorf("'%s' is not a string", timeStampStr)
	}
	tim, err := parseTime(fmt.Sprintf("%v", ts))
	if err != nil {
		return nil, fmt.Errorf("'%s' could not be parsed", timeStampStr)
	}
	_, ok = m[groupsStr]
	if !ok {
		return nil, fmt.Errorf("'%s' does not exist in data root", groupsStr)
	}

	dr := &DataRoot{timeStamp: tim, dataMap: m, navIndex: createNavIndex(m), dataMapUpdated: dataMapUpdated}
	return dr, nil
}

func (p *DataRoot) Search(addPath func(string), needle string, matchCase bool) {
	rootMap := p.dataMap
	groups := rootMap[groupsStr].(map[string]interface{})
	for user, v := range groups {
		searchUsers(addPath, needle, user, v.(map[string]interface{}), matchCase)
	}
}

func searchUsers(addPath func(string), needle, user string, m map[string]interface{}, matchCase bool) {
	for k, v := range m {
		if k == hintStr {
			for k1, v1 := range v.(map[string]interface{}) {
				searchLeafNodes(addPath, hintStr, needle, user, k1, v1.(map[string]interface{}), matchCase)
			}
		} else {
			searchLeafNodes(addPath, "", needle, user, k, v.(map[string]interface{}), matchCase)
		}
	}
}

func searchLeafNodes(addPath func(string), tag, needle, user, name string, m map[string]interface{}, matchCase bool) {
	if tag != "" {
		tag = "." + tag + "."
	} else {
		tag = "."
	}
	if containsWithCase(name, needle, matchCase) {
		addPath(user + tag + name)
	}
	for k, s := range m {
		if containsWithCase(k, needle, matchCase) {
			addPath(user + tag + name)
		}
		if reflect.ValueOf(s).Kind() == reflect.String {
			if containsWithCase(s.(string), needle, matchCase) {
				addPath(user + tag + name)
			}
		} else {
			searchLeafNodes(addPath, name, needle, user, k, s.(map[string]interface{}), matchCase)
		}
	}
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

func (p *DataRoot) AddUser(userName string) error {
	m := GetMapForUid(userName, &p.dataMap)
	if m != nil {
		return fmt.Errorf("user name '%s' already exists", userName)
	}
	rootMap := p.dataMap
	groups := rootMap[groupsStr].(map[string]interface{})

	newUser := make(map[string]interface{})
	addHint("application", userName, newUser)
	addNote(userName, "note", newUser)

	groups[userName] = newUser
	p.navIndex = createNavIndex(p.dataMap)
	p.dataMapUpdated("Added user:", userName, userName, nil)
	return nil
}

func (p *DataRoot) AddNote(userName, noteName string) error {
	user := GetMapForUid(userName, &p.dataMap)
	if user == nil {
		return fmt.Errorf("user name '%s' does not exists", userName)
	}
	path, err := addNote(userName, noteName, *user)
	p.navIndex = createNavIndex(p.dataMap)
	p.dataMapUpdated("Added note:", userName, path, err)
	return err
}

func (p *DataRoot) AddHint(userName, appName string) error {
	user := GetMapForUid(userName, &p.dataMap)
	if user == nil {
		return fmt.Errorf("user name '%s' does not exists", userName)
	}
	path, err := addHint(appName, userName, *user)
	p.navIndex = createNavIndex(p.dataMap)
	p.dataMapUpdated("Added hint:", userName, path, err)
	return err
}

func addNote(userName, noteName string, user map[string]interface{}) (string, error) {
	_, ok := user[noteStr]
	if !ok {
		user[noteStr] = make(map[string]interface{})
	}
	notes := user[noteStr].(map[string]interface{})
	_, ok = notes[noteName]
	if !ok {
		notes[noteName] = ""
		return fmt.Sprintf("%s.%s", userName, noteStr), nil
	}
	return "", fmt.Errorf("note '%s' already exists", noteName)
}

func addHint(hintName, userName string, user map[string]interface{}) (string, error) {
	_, ok := user[hintStr]
	if !ok {
		user[hintStr] = make(map[string]interface{})
	}
	hints := user[hintStr].(map[string]interface{})
	_, ok = hints[hintName]
	if !ok {
		hints[hintName] = make(map[string]interface{})
		hint := hints[hintName].(map[string]interface{})
		hint["userId"] = ""
		hint["pre"] = ""
		hint["post"] = ""
		hint["notes"] = ""
		hint["positional"] = "12345"
		return fmt.Sprintf("%s.%s.%s", userName, hintStr, hintName), nil
	}
	return "", fmt.Errorf("hint '%s' already exists", hintName)
}

func (p *DataRoot) GetRootUidOrCurrentUid(currentUid string) string {
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

func (r *DataRoot) GetDataRootMap() *map[string]interface{} {
	return &r.dataMap
}

func (r *DataRoot) GetTimeStamp() time.Time {
	return r.timeStamp
}

func (r *DataRoot) GetNavIndex(id string) []string {
	return r.navIndex[id]
}

func (r *DataRoot) ToJson() (string, error) {
	output, err := json.MarshalIndent(r.dataMap, "", "    ")
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func (r *DataRoot) ToJsonTreeMap() string {
	output, err := json.MarshalIndent(r.navIndex, "", "    ")
	if err != nil {
		return err.Error()
	}
	return string(output)
}

func (r *DataRoot) ToStruct() string {
	var sb strings.Builder
	appendMapStruct(&sb, r.dataMap, 1)
	return sb.String()
}

func GetParentId(uid string) string {
	if uid == "" {
		return uid
	}
	p := strings.LastIndexByte(uid, '.')
	switch p {
	case -1:
		return uid
	case 0:
		return ""
	default:
		return uid[0:p]
	}
}

func GetUserFromPath(path string) string {
	return GetFirstPathElements(path, 1)
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

func GetMapForUid(uid string, m *map[string]interface{}) *map[string]interface{} {
	nodes := strings.Split(uid, ".")
	if len(nodes) == 1 && nodes[0] == "" {
		return m
	}
	n := *m
	x := n[groupsStr]
	for _, v := range nodes {
		y := x.(map[string]interface{})[v]
		if y == nil {
			return nil
		}
		if reflect.ValueOf(y).Kind() != reflect.String {
			x = y
		}
	}
	o := x.(map[string]interface{})
	return &o
}

func appendMapStruct(sb *strings.Builder, m map[string]interface{}, ind int) {
	for k, v := range m {
		if reflect.ValueOf(v).Kind() != reflect.String {
			sb.WriteString(fmt.Sprintf("%d:%s:%s \n", ind, tabdata[:ind*2], k))
			appendMapStruct(sb, v.(map[string]interface{}), ind+1)
		} else {
			sb.WriteString(fmt.Sprintf("%d:%s-%s = %s\n", ind, tabdata[:ind*2], k, v))
		}
	}
}

func createNavIndex(m map[string]interface{}) map[string][]string {
	var uids = make(map[string][]string)
	createNavIndexDetail("", uids, m)
	return uids
}

func createNavIndexDetail(id string, uids map[string][]string, m map[string]interface{}) {
	for _, v := range m {
		if reflect.ValueOf(v).Kind() != reflect.String {
			l, ll := keysToList(id, v.(map[string]interface{}))
			uids[id] = ll
			for ii2, id2 := range l {
				m2 := v.(map[string]interface{})[id2]
				for _, v2 := range m2.(map[string]interface{}) {
					if reflect.ValueOf(v2).Kind() != reflect.String {
						l2, ll2 := keysToList(ll[ii2], m2.(map[string]interface{}))
						uids[ll[ii2]] = ll2
						for ii3, id3 := range l2 {
							m3 := m2.(map[string]interface{})[id3]
							for _, v3 := range m3.(map[string]interface{}) {
								if reflect.ValueOf(v3).Kind() != reflect.String {
									_, ll3 := keysToList(ll2[ii3], m3.(map[string]interface{}))
									uids[ll2[ii3]] = ll3
								}
							}
						}
					}
				}
			}
		}
	}
}

func keysToList(id string, m map[string]interface{}) ([]string, []string) {
	l := make([]string, 0)
	ll := make([]string, 0)
	for k := range m {
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

func parseJson(j []byte) (map[string]interface{}, error) {
	var m map[string]interface{}
	err := json.Unmarshal([]byte(j), &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}
