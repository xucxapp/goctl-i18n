package main

import (
	"fmt"
	"os"

	localplugin "github.com/xucxapp/goctl-i18n/internal/plugin"
)

// main 是 goctl-i18n 的程序入口。
func main() {
	if err := localplugin.Execute(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
