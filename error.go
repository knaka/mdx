package mdx

import "fmt"

type MdxError struct {
	msg      string
	absPath  string
	source   []byte
	position int
}

func (me *MdxError) Error() string {
	lineNo := 1
	for i := me.position; i >= 0; i-- {
		r := rune(me.source[i])
		if r == '\n' {
			lineNo++
		}
	}
	return fmt.Sprintf("%s (%s:%d)", me.msg, me.absPath, lineNo)
}

var _ error = (*MdxError)(nil)

func NewError(msg string, absPath string, source []byte, position int) *MdxError {
	return &MdxError{msg, absPath, source, position}
}
