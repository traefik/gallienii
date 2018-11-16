package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/BurntSushi/toml"
	"github.com/containous/flaeg"
	"github.com/containous/gallienii/generate"
	"github.com/containous/gallienii/meta"
	"github.com/containous/gallienii/sync"
	"github.com/containous/gallienii/types"
	"github.com/google/go-github/github"
	"github.com/ogier/pflag"
	"golang.org/x/oauth2"
)

func main() {

	// Root

	emptyConfig := &types.NoOption{}
	rootCmd := &flaeg.Command{
		Name:                  "gallienii",
		Description:           `Myrmica gallienii: Keep forks synchronized.`,
		Config:                emptyConfig,
		DefaultPointersConfig: &types.NoOption{},
		Run: func() error {
			// no-op
			return nil
		},
	}

	flag := flaeg.New(rootCmd, os.Args[1:])

	// Sync

	syncOptions := &types.SyncOptions{
		DryRun:         true,
		ConfigFilePath: "./gallienii.toml",
		ServerPort:     80,
	}

	syncCmd := &flaeg.Command{
		Name:        "sync",
		Description: "Synchronize forks.",
		Config:      syncOptions,
		DefaultPointersConfig: &types.SyncOptions{
			DryRun: true,
		},
		Run: runSync(syncOptions),
	}

	flag.AddCommand(syncCmd)

	// Generate

	generateOptions := &types.GenerateOptions{}

	generateCmd := &flaeg.Command{
		DefaultPointersConfig: &types.GenerateOptions{},
		Description:           "Generate configuration file.",
		Name:                  "gen",
		Config:                generateOptions,
		Run:                   runGenerate(generateOptions),
	}

	flag.AddCommand(generateCmd)

	// version

	versionOptions := &types.NoOption{}

	versionCmd := &flaeg.Command{
		Name:                  "version",
		Description:           "Display the version.",
		Config:                versionOptions,
		DefaultPointersConfig: &types.NoOption{},
		Run: func() error {
			meta.DisplayVersion()
			return nil
		},
	}

	flag.AddCommand(versionCmd)

	// Run command
	err := flag.Run()
	if err != nil && err != pflag.ErrHelp {
		log.Fatalf("Error: %v\n", err)
	}
}

func runSync(options *types.SyncOptions) func() error {
	return func() error {
		err := required(options.ConfigFilePath, "config-path")
		if err != nil {
			return err
		}

		if len(options.GitHubToken) == 0 {
			options.GitHubToken = os.Getenv("GITHUB_TOKEN")
		}

		if options.DryRun {
			log.Print("IMPORTANT: you are using the dry-run mode. Use `--dry-run=false` to disable this mode.")
		}

		configs, err := readConfiguration(options.ConfigFilePath)
		if err != nil {
			return err
		}

		if options.Verbose {
			log.Printf("%+v", configs)
		}

		if options.ServerMode {
			server := &server{options: options, configs: configs}
			return server.ListenAndServe()
		}

		ctx := context.Background()
		client := NewGitHubClient(ctx, options.GitHubToken)

		return sync.Process(ctx, client, configs, options.DryRun, options.Verbose)
	}
}

func readConfiguration(path string) (*types.Configuration, error) {
	config := &types.Configuration{}
	_, err := toml.DecodeFile(path, config)

	if err != nil {
		return nil, err
	}

	return config, nil
}

func runGenerate(options *types.GenerateOptions) func() error {
	return func() error {
		ctx := context.Background()
		client := NewGitHubClient(ctx, options.GitHubToken)

		switch {
		case options.Sample:
			err := generate.Sample("./sample.toml")
			if err != nil {
				return err
			}
		case options.User != "":
			err := generate.UserConfiguration(ctx, client, options.User, "./gallienii.toml")
			if err != nil {
				return err
			}
		case options.Org != "":
			err := generate.OrganizationConfiguration(ctx, client, options.Org, "./gallienii.toml")
			if err != nil {
				return err
			}
		default:
			return errors.New("one option must be fill")
		}
		return nil
	}
}

// NewGitHubClient create a new GitHub client
func NewGitHubClient(ctx context.Context, token string) *github.Client {
	var client *github.Client
	if len(token) == 0 {
		client = github.NewClient(nil)
	} else {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		tc := oauth2.NewClient(ctx, ts)
		client = github.NewClient(tc)
	}
	return client
}

type server struct {
	options *types.SyncOptions
	configs *types.Configuration
}

func (s *server) ListenAndServe() error {
	return http.ListenAndServe(":"+strconv.Itoa(s.options.ServerPort), s)
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		log.Printf("Invalid http method: %s", r.Method)
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	ctx := context.Background()
	client := NewGitHubClient(ctx, s.options.GitHubToken)
	err := sync.Process(ctx, client, s.configs, s.options.DryRun, s.options.Verbose)

	if err != nil {
		log.Printf("Sync error: %v", err)
		http.Error(w, "Sync error.", http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, "Myrmica gallienii: Scheluded.\n")
}

func required(field string, fieldName string) error {
	if len(field) == 0 {
		return fmt.Errorf("%s is mandatory", fieldName)
	}
	return nil
}
