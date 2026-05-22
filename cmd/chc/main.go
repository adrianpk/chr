package main

import (
	"context"
	"embed"
	"fmt"
	"os"

	"forge.adrianpk.com/chc/internal/chcapp"
)

//go:embed assets
var assetsFS embed.FS

func main() {
	err := chcapp.Run(context.Background(), os.Args[1:], os.Stdin, os.Stdout, os.Stderr, assetsFS)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
