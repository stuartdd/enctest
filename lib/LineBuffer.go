package lib

type Line struct {
	chars []rune
	pos   int
	max   int
}

func NewLine(max int) *Line {
	l := &Line{chars: make([]rune, max), max: max, pos: 0}
	l.Clear()
	return l
}

func (l *Line) Clear() {
	for i := 0; i < l.max; i++ {
		l.chars[i] = ' '
	}
	l.pos = 0
}

func (l *Line) Apply(s string, len int) {
	for i, c := range s {
		l.chars[l.pos+i] = c
	}
	l.pos = l.pos + (len - 1)
}

func (l *Line) String() string {
	return string(l.chars[0:l.pos])
}
