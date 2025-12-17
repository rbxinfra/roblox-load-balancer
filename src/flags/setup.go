package flags

import (
	"flag"
	"fmt"
	"os"
	"sync"
)

var setupOnce sync.Once

// SetupFlags sets up flags for the current environment.
func SetupFlags(applicationName, buildMode, commitSha string) {
	setupOnce.Do(func() {
		flag.Usage = func() {
			fmt.Fprintf(os.Stderr, "Usage: %s\nBuild Mode: %s\nCommit: %s %s\n\n", applicationName, buildMode, commitSha, FlagsUsageString)
			flag.PrintDefaults()
		}

		flag.Set("logtostderr", "true")
		flag.Set("v", "50")

		flag.Parse()
	})
}
