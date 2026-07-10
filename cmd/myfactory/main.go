// myfactory is the standalone MyFactory CLI.
// All reusable project templates are embedded in the binary; no Python, Go,
// or source checkout is required at runtime.
package main

import (
	"os"

	"github.com/AlharbiKhalid/myfactory-core/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:], os.Stdout, os.Stderr))
}
