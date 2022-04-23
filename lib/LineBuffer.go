package lib

import "strings"

type Line struct {
	chars []rune
	pos   int
	max   int
}

func NewLine(max int) *Line {
	l := &Line{chars: make([]rune, max), max: max - 1, pos: 0}
	l.Clear()
	return l
}

func (l *Line) Clear() {
	for i := 0; i < (l.max + 1); i++ {
		l.chars[i] = ' '
	}
	l.pos = 0
}

func (l *Line) Apply(s string, width int) {
	pos := l.pos
	for i, c := range s {
		if pos+i <= (l.max) {
			l.chars[pos+i] = c
		}
	}
	l.pos = pos + width
}

func (l *Line) ApplyRev(s string, width int) {
	pos := l.pos + width - len(s)
	for i, c := range s {
		if pos+i <= (l.max) {
			l.chars[pos+i] = c
		}
	}
	l.pos = l.pos + width
}

func (l *Line) Contains(s string) bool {
	if s == "" {
		return true
	}
	return strings.Contains(l.String(), s)
}

func (l *Line) String() string {
	if l.pos <= l.max {
		return string(l.chars[0:l.pos])
	}
	return string(l.chars[0 : l.max+1])
}
