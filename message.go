package main

import (
	"fmt"
	"regexp"
	"strings"
)

type Message struct {
	MCUser *MCUser
	Text   string

	Output func() string
}

func NewMessage(t string) *Message {
	var err error
	m := new(Message)
	msg_user := regexp.MustCompile("^<[^>]+>")
	tmpMCUser := msg_user.FindString(t)
	tmpMCUser = strings.Replace(tmpMCUser, "<", "", -1)
	tmpMCUser = strings.Replace(tmpMCUser, ">", "", -1)
	if tmpMCUser != "" {
		if m.MCUser, err = c.model.getMCUser(tmpMCUser); err != nil {
			m.MCUser = new(MCUser)
			m.MCUser.Name = tmpMCUser
			fmt.Println(">>> Updating/Creating User: " + m.MCUser.Name)
			c.model.updateMCUser(m.MCUser)
		}
	}

	m.Text = t

	if m.MCUser != nil {
		res := strings.Split(t, "<"+m.MCUser.Name+"> ")
		if len(res) > 0 {
			m.Text = res[1]
		}
	}

	m.Output = func() string {
		if m.MCUser != nil {
			return "<" + m.MCUser.Name + "> " + m.Text
		}
		return m.Text
	}

	return m
}
