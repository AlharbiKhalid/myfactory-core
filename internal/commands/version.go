// Command version: print build metadata.
package commands

import (
	"fmt"
	"io"

	"github.com/AlharbiKhalid/myfactory-core/internal/version"
)

// Version implements `myfactory version`.
func Version(args []string, stdout, stderr io.Writer) int {
	fmt.Fprintf(stdout, "MyFactory version: %s\n", version.Version)
	fmt.Fprintf(stdout, "Git commit:        %s\n", version.Commit)
	fmt.Fprintf(stdout, "Build date:        %s\n", version.Date)
	return 0
}
