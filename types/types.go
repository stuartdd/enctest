package types

import (
	"strings"
)

type NodeAnnotationEnum int

const (
	NOTE_TYPE_SL NodeAnnotationEnum = iota
	NOTE_TYPE_ML
	NOTE_TYPE_RT
	NOTE_TYPE_PO
)

var (
	nodeAnnotationPrefix      = []string{"", "ml!", "rt!", "po!"}
	NodeAnnotationPrefixNames = []string{"Single Line", "Multi Line", "Rich Text", "Positional"}
	NodeAnnotationEnums       = []NodeAnnotationEnum{NOTE_TYPE_SL, NOTE_TYPE_ML, NOTE_TYPE_RT, NOTE_TYPE_PO}
)

func GetNodeName(combinedName string) string {
	_, n := GetNodeAnnotationTypeAndName(combinedName)
	return n
}

func GetNodeAnnotationTypeAndName(combinedName string) (NodeAnnotationEnum, string) {
	if strings.HasPrefix(combinedName, nodeAnnotationPrefix[NOTE_TYPE_ML]) {
		return NOTE_TYPE_ML, combinedName[3:]
	}
	if strings.HasPrefix(combinedName, nodeAnnotationPrefix[NOTE_TYPE_RT]) {
		return NOTE_TYPE_RT, combinedName[3:]
	}
	if strings.HasPrefix(combinedName, nodeAnnotationPrefix[NOTE_TYPE_PO]) {
		return NOTE_TYPE_PO, combinedName[3:]
	}
	return NOTE_TYPE_SL, combinedName
}

func GetNodeAnnotationNameWithPrefix(nae NodeAnnotationEnum, name string) string {
	return nodeAnnotationPrefix[nae] + name
}

func GetNodeAnnotationPrefixName(nae NodeAnnotationEnum) string {
	return nodeAnnotationPrefix[nae]
}
