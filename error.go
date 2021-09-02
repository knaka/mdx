package mdpp

import "fmt"

type MdppError struct {
	msg      string
	absPath  string
	source   []byte
	position int
}

func (me *MdppError) Error() string {
	lineNo := 1
	for i := me.position; i >= 0; i-- {
		r := rune(me.source[i])
		if r == '\n' {
			lineNo++
		}
	}
	return fmt.Sprintf("%s (%s:%d)", me.msg, me.absPath, lineNo)
}

var _ error = (*MdppError)(nil)

func NewError(msg string, absPath string, source []byte, position int) *MdppError {
	return &MdppError{msg, absPath, source, position}
}
