package main

import (
	"fmt"
	"os"
	"time"

	"github.com/hiddenpathz/changeloger/internal/changeloger"
)

func main() {
	app := changeloger.NewApp(os.Args[1:], os.Stdin, os.Stdout, time.Now, changeloger.OSRunner{})

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "\033[01;31mОшибка: %v\n\033[0m", err)
		os.Exit(1)
	}
}
