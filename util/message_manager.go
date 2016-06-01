package util

import (
	"fmt"
	"io"
	"strings"
)

type MessageManager struct {
	/* listeners is an array of functions that
	 * tell the manager how to listen for specific
	 * text and what to do if we receive it.
	 * Each listener returns true if the input was
	 * "consumed" (i.e. - Don't send to any more
	 * listeners)
	 */
	listeners []func(inp *Message) bool

	/* tempListeners is an array of functions that
	 * work the same as 'listeners', but these are
	 * just temporary and higher priority than
	 * 'listeners'
	 */
	tempListeners []func(inp *Message) bool

	/* finishedTempListener is an index of a
	 * tempListener to be removed
	 */
	finishedTempListener int

	output io.WriteCloser
}

var Listeners []func(inp *Message) bool
var TempListeners []func(inp *Message) bool

func (mm MessageManager) Tell(user string, what string, color string) {
	valid_color := false
	switch color {
	case "black", "dark_blue", "dark_green", "dark_aqua",
		"dark_red", "dark_purple", "gold", "gray", "dark_gray",
		"blue", "green", "aqua", "red", "light_purple", "yellow",
		"white":
		valid_color = true
	}
	if !valid_color {
		color = "white"
	}
	strings.Replace(what, "\n", "", -1)
	o := "tellraw " + user + " {text:\"" + what + "\",color:" + color + "}"
	mm.Output(o)
}

func (mm MessageManager) TellRaw(user string, what string) {
	o := "tellraw " + user + " {text:\"" + what + "\"}"
	mm.Output(o)
}

func (mm MessageManager) Output(o string) {
	if !strings.HasSuffix("\n", o) {
		o = o + "\n"
	}
	mm.output.Write([]byte(o))
	fmt.Printf("%s", o)
}

func (mm MessageManager) ProcessMessage(inp string) bool {
	// First of all, create the message from inp
	m := NewMessage(inp)
	// Now run the message through all of mm's tempListeners
	for i := range TempListeners {
		// Pop the listener off of the stack
		consumed := TempListeners[i](m)
		if consumed {
			// When a temp listener is consumed, we delete it
			RemoveTempListener(i)
			return true
		}
	}
	// and run the message through all of mm's listeners
	for i := range Listeners {
		consumed := Listeners[i](m)
		if consumed {
			return true
		}
	}
	// Message not consumed, return false
	return false
}

func AddListener(l func(*Message) bool) {
	Listeners = append(Listeners, l)
}

func AddTempListener(l func(*Message) bool) {
	TempListeners = append(TempListeners, l)
}

func RemoveListener(i int) {
	if i > -1 && i < len(Listeners) {
		t := append(Listeners[:i], Listeners[i+1:]...)
		Listeners = make(([]func(*Message) bool), len(t))
		copy(Listeners, t)
	}
}
func RemoveTempListener(i int) {
	if i > -1 && i < len(TempListeners) {
		t := append(TempListeners[:i], TempListeners[i+1:]...)
		TempListeners = make(([]func(*Message) bool), len(t))
		copy(TempListeners, t)
	}
}

func NewManager(o io.WriteCloser) MessageManager {
	mm := MessageManager{
		//		listeners:     make(([]func(*Message) bool), 0, 10),
		//		tempListeners: make(([]func(*Message) bool), 0, 10),
		output: o,
	}
	return mm
}
