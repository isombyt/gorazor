package helper

import (
	"bytes"
	"github.com/sipin/gorazor/gorazor"
	. "kp/models"
)

func Msg(u *User) string {
	var _buffer bytes.Buffer
	_buffer.WriteString("\n\n\n")
	{
		username := u.Name
		if u.Email != "" {
			username += "(" + u.Email + ")"
		}
	}
	_buffer.WriteString("\n<div class=\"welcome\">\n<h4>Hello ")
	_buffer.WriteString(gorazor.HTMLEscape(username))
	_buffer.WriteString("</h4>\n\n<div>")
	_buffer.WriteString((u.Intro))
	_buffer.WriteString("</div>\n</div>\n\n")

	return _buffer.String()
}
