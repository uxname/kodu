// Command kodu — CLI для подготовки кодовой базы под LLM.
package main

import (
	"fmt"
	"os"

	"github.com/uxname/kodu/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
