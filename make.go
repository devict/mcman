// This file just holds "go generate" commands

//go:generate esc -o static.go public tmpls
//go:generate sh -c "godoc2md $PWD > README.md"

package main
