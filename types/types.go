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
	nodeAnnotationPrefix      = []string{"", "!ml", "!rt", "!po"}
	NodeAnnotationPrefixNames = []string{"Single Line", "Multi Line", "Rich Text", "Positional"}
	NodeAnnotationEnums       = []NodeAnnotationEnum{NOTE_TYPE_SL, NOTE_TYPE_ML, NOTE_TYPE_RT, NOTE_TYPE_PO}
)

func IndexOfAnnotation(annotation string) int {
	for i, v := range nodeAnnotationPrefix {
		if v == annotation {
			return i
		}
	}
	return 0
}

func GetNodeAnnotationTypeAndName(combinedName string) (NodeAnnotationEnum, string) {
	pos := strings.IndexRune(combinedName, '!')
	if pos < 0 {
		return NOTE_TYPE_SL, combinedName
	}
	aStr := combinedName[pos:]
	indx := IndexOfAnnotation(aStr)
	if indx == 0 {
		return NOTE_TYPE_SL, combinedName
	}
	return NodeAnnotationEnums[indx], combinedName[:pos]
}

func GetNodeAnnotationNameWithPrefix(nae NodeAnnotationEnum, name string) string {
	return name + nodeAnnotationPrefix[nae]
}

func GetNodeAnnotationPrefixName(nae NodeAnnotationEnum) string {
	return nodeAnnotationPrefix[nae]
}
