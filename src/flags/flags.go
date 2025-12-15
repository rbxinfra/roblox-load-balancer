package flags

import "flag"

var (
	// HelpFlag prints the usage.
	HelpFlag = flag.Bool("help", false, "Print usage.")

	// ConfigurationFilePath is the path to the static configuration.
	ConfigurationFilePath = flag.String("configuration-file-path", "", "The path to the static configuration.")

	// DryRun reads from Consul, builds the config, and outputs to the file without starting the Daemon or reloading HAProxy.
	DryRun = flag.Bool("dry-run", false, "Reads from Consul, builds the config, and outputs to the file without starting the Daemon or reloading HAProxy.")
)

const FlagsUsageString string = `
	[-h|--help]
	[--configuration-file-path[=]] [--dry-run]`
