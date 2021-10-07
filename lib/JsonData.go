package lib

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/stuartdd/jsonParserGo/parser"
	"stuartdd.com/types"
)

const (
	hintStr            = "pwHints"
	noteStr            = "notes"
	dataMapRootName    = "groups"
	timeStampStr       = "timeStamp"
	tabdata            = "                                     "
	allowedCharsInName = " *@#$%^&*()_+=?"
)

var (
	defaultHintNames = []string{"notes", "post", "pre", "userId"}
)

type JsonData struct {
	timeStamp      time.Time
	dataMap        *parser.JsonObject
	navIndex       map[string][]string
	dataMapUpdated func(string, string, string, error)
}

func NewJsonData(j []byte, dataMapUpdated func(string, string, string, error)) (*JsonData, error) {
	mIn, err := parser.Parse(j)
	if err != nil {
		return nil, err
	}
	if mIn.GetNodeType() != parser.NT_OBJECT {
		return nil, fmt.Errorf("root element is NOT a JsonObject")
	}
	rO := mIn.(*parser.JsonObject)

	u := rO.GetNodeWithName(dataMapRootName)
	if u.GetNodeType() != parser.NT_OBJECT {
		return nil, fmt.Errorf("root '%s' element is NOT a JsonObject", dataMapRootName)
	}

	ts, err := parser.Find(rO, timeStampStr)
	if err != nil {
		return nil, fmt.Errorf("'%s' does not exist in data root", timeStampStr)
	}
	tim, err := parseTime((ts.(*parser.JsonString)).GetValue())
	if err != nil {
		return nil, fmt.Errorf("'%s' could not be parsed", timeStampStr)
	}
	dr := &JsonData{timeStamp: tim, dataMap: rO, navIndex: *createNavIndex(rO), dataMapUpdated: dataMapUpdated}
	return dr, nil
}

func (p *JsonData) GetTimeStamp() time.Time {
	return p.timeStamp
}

func (p *JsonData) GetNavIndex(id string) []string {
	return p.navIndex[id]
}

func (p *JsonData) GetDataRoot() *parser.JsonObject {
	return p.dataMap
}

/*
	We know we can cast the groups node. We checked it in NewJsonData
*/
func (p *JsonData) GetUserRoot() *parser.JsonObject {
	return p.dataMap.GetNodeWithName(dataMapRootName).(*parser.JsonObject)
}

func (p *JsonData) GetUserNode(user string) *parser.JsonObject {
	u := p.GetUserRoot().GetNodeWithName(user)
	if u == nil {
		return nil
	}
	return u.(*parser.JsonObject)
}

func (p *JsonData) ToJson() string {
	return p.dataMap.String()
}

func (p *JsonData) Search(addPath func(string, string), needle string, matchCase bool) {
	groups := p.dataMap.GetNodeWithName(dataMapRootName).(*parser.JsonObject)
	for _, v := range groups.GetValues() {
		searchUsers(addPath, needle, v.GetName(), v.(*parser.JsonObject), matchCase)
	}
}

func (p *JsonData) AddHintItem(user, path, hintItemName string) error {
	h, err := parser.Find(p.GetUserRoot(), path)
	if err != nil {
		return fmt.Errorf("the hint '%s' cannot be found", path)
	}
	if h.GetNodeType() != parser.NT_OBJECT {
		return fmt.Errorf("the hint note '%s' was not an object node", path)
	}
	hO := h.(*parser.JsonObject)
	addStringIfDoesNotExist(hO, hintItemName)
	p.navIndex = *createNavIndex(p.dataMap)
	p.dataMapUpdated("Add Note Item", GetUserFromPath(user), path, nil)
	return nil
}

func (p *JsonData) AddHint(user, hintName string) error {
	u := p.GetUserNode(user)
	if u == nil {
		return fmt.Errorf("the user '%s' cannot be found", user)
	}
	addHintToUser(u, hintName)
	p.navIndex = *createNavIndex(p.dataMap)
	p.dataMapUpdated("Add Note Item", GetUserFromPath(user), user+"."+hintStr+"."+hintName, nil)
	return nil
}

func (p *JsonData) AddNoteItem(user, itemName string) error {
	u := p.GetUserNode(user)
	if u == nil {
		return fmt.Errorf("the user '%s' cannot be found", user)
	}
	addNoteToUser(u, itemName, "")
	p.navIndex = *createNavIndex(p.dataMap)
	p.dataMapUpdated("Add Note Item", GetUserFromPath(user), user+"."+noteStr, nil)
	return nil
}

func (p *JsonData) AddUser(user string) error {
	u := p.GetUserNode(user)
	if u != nil {
		return fmt.Errorf("the user '%s' already exists", user)
	}
	userO := parser.NewJsonObject(user)
	addNoteToUser(userO, "note", "text")
	addHintToUser(userO, "App1")
	p.GetUserRoot().Add(userO)
	p.navIndex = *createNavIndex(p.dataMap)
	p.dataMapUpdated("New User", GetUserFromPath(user), user, nil)
	return nil
}

func (p *JsonData) Rename(uid string, newName string) error {
	n, err := p.GetUserDataForUid(uid)
	if err != nil {
		return fmt.Errorf("the item to rename '%s' was not found in the data", uid)
	}
	parent, ok := parser.FindParentNode(p.dataMap, n)
	if !ok {
		return fmt.Errorf("the item to rename '%s' does not have a valid parent", uid)
	}
	err = parser.Rename(p.dataMap, n, newName)
	p.navIndex = *createNavIndex(p.dataMap)

	if parent.GetName() == dataMapRootName { // If the parent is groups then the user was renamed
		p.dataMapUpdated("Rename", GetUserFromPath(newName), newName, nil)
	} else {
		p.dataMapUpdated("Rename", GetUserFromPath(uid), GetParentId(uid)+"."+newName, nil)
	}
	return nil
}

func (p *JsonData) Remove(uid string, min int) error {
	n, err := p.GetUserDataForUid(uid)
	if err != nil {
		return fmt.Errorf("the item to remove '%s' was not found in the data", uid)
	}
	parent, ok := parser.FindParentNode(p.dataMap, n)
	if !ok {
		return fmt.Errorf("the item to remove '%s' does not have a valid parent", uid)
	}
	if parent.GetNodeType() != parser.NT_OBJECT {
		return fmt.Errorf("the item to remove '%s' does not have an object parent", uid)
	}
	parentObj := parent.(*parser.JsonObject)
	count := parentObj.Len()
	if count <= min {
		return fmt.Errorf("there must be at least %d element(s) remaining in this item", min)
	}
	parser.Remove(p.dataMap, n)
	p.navIndex = *createNavIndex(p.dataMap)
	p.dataMapUpdated("Removed", GetUserFromPath(uid), GetParentId(uid), nil)
	return nil
}

func (p *JsonData) IsStringNode(n parser.NodeI) bool {
	return n.GetNodeType() == parser.NT_STRING
}

func (p *JsonData) GetUserDataForUid(uid string) (parser.NodeI, error) {
	return GetUserDataForUid(p.GetDataRoot(), uid)
}

func GetUserDataForUid(root parser.NodeI, uid string) (parser.NodeI, error) {
	nodes, err := parser.Find(root, dataMapRootName+"."+uid)
	if err != nil {
		return nil, err
	}
	return nodes, nil
}

/**
Validate the names of entities. These result in JSON entity names so require
some restrictions.
*/
func ProcessEntityName(entry string, nt types.NodeAnnotationEnum) (string, error) {
	if len(entry) == 0 {
		return "", fmt.Errorf("input is undefined")
	}
	if len(entry) < 2 {
		return "", fmt.Errorf("input '%s' is too short. Must be longer that 1 char", entry)
	}
	lcEntry := strings.ToLower(entry)
	for _, c := range lcEntry {
		if c < ' ' {
			return "", fmt.Errorf("input cannot not contain control characters")
		}
		if (c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (strings.ContainsRune(allowedCharsInName, c)) {
			continue
		}
		return "", fmt.Errorf("input must not contain character '%c'. Only '0..9', 'a..z', 'A..Z' and '%s' chars are allowed", c, allowedCharsInName)
	}
	return types.GetNodeAnnotationNameWithPrefix(nt, entry), nil
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

func addDefaultHintItemsToHint(hint *parser.JsonObject) {
	for _, n := range defaultHintNames {
		addStringIfDoesNotExist(hint, n)
	}
}

func addStringIfDoesNotExist(obj *parser.JsonObject, name string) {
	node := obj.GetNodeWithName(name)
	if node == nil {
		node = parser.NewJsonString(name, "")
		obj.Add(node)
	}
}

func addHintToUser(userO *parser.JsonObject, hintName string) {
	pwHints := userO.GetNodeWithName(hintStr)
	if pwHints == nil {
		pwHints = parser.NewJsonObject(hintStr)
		userO.Add(pwHints)
	}
	pwHintsO := pwHints.(*parser.JsonObject)
	hint := pwHintsO.GetNodeWithName(hintName)
	if hint == nil {
		hint = parser.NewJsonObject(hintName)
		pwHintsO.Add(hint)
	}
	hintO := hint.(*parser.JsonObject)
	addDefaultHintItemsToHint(hintO)
}

func addNoteToUser(userO *parser.JsonObject, noteName, noteText string) {
	notes := userO.GetNodeWithName(noteStr)
	if notes == nil {
		notes = parser.NewJsonObject(noteStr)
		userO.Add(notes)
	}
	notesO := notes.(*parser.JsonObject)
	note := notesO.GetNodeWithName(noteName)
	if note == nil {
		note = parser.NewJsonString(noteName, noteText)
		notesO.Add(note)
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
		return uid
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
