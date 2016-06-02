// This file just holds "go generate" commands

//go:generate esc -o static.go public tmpls
//go:generate sh -c "godoc2md github.com/devict/mcman > README.md"

package main
