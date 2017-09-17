package types

// NoOption empty struct.
type NoOption struct{}

// GenerateOptions Generator command options.
type GenerateOptions struct {
	Sample      bool   `description:"Generate a sample configuration file."`
	Org         string `description:"Generate a default configuration file for an organization name."`
	User        string `description:"Generate a default configuration file for a user name."`
	GitHubToken string `long:"token" short:"t" description:"GitHub Token."`
}

// SyncOptions Synchronizer command options.
type SyncOptions struct {
	GitHubToken   string `long:"token" short:"t" description:"GitHub Token [required]."`
	RulesFilePath string `long:"rules-path" description:"Path to the rules file."`
	ServerMode    bool   `long:"server" description:"Server mode."`
	ServerPort    int    `long:"port" description:"Server port."`
	DryRun        bool   `long:"dry-run" description:"Dry run mode."`
	Verbose       bool   `long:"verbose" description:"Verbose mode."`
}
