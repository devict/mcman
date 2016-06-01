package util

import (
	"regexp"
	"strings"
)

type Message struct {
	MCUser *MCUser
	Text   string

	Output func() string
}

func NewMessage(t string) *Message {
	m := new(Message)
	msg_user := regexp.MustCompile("<[^>]+>")
	tmpMCUser := msg_user.FindString(t)
	tmpMCUser = strings.Replace(tmpMCUser, "<", "", -1)
	tmpMCUser = strings.Replace(tmpMCUser, ">", "", -1)
	m.MCUser = FindMCUser(tmpMCUser, true)
	m.Text = t
	if m.MCUser.Index != -1 && m.MCUser.Name != "" {
		res := strings.Split(t, "<"+m.MCUser.Name+"> ")
		if len(res) > 0 {
			m.Text = res[1]
		}
	}

	m.Output = func() string {
		if m.MCUser.Index != -1 && m.MCUser.Name != "" {
			return "<" + m.MCUser.Name + "> " + m.Text
		} else {
			return m.Text
		}
	}

	return m
}
