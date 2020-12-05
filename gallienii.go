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
	"github.com/google/go-github/v27/github"
	"github.com/ogier/pflag"
	"github.com/traefik/gallienii/generate"
	"github.com/traefik/gallienii/sync"
	"github.com/traefik/gallienii/types"
	"golang.org/x/oauth2"
)

func main() {
	// Root

	rootCmd := &flaeg.Command{
		Name:                  "gallienii",
		Description:           `Myrmica gallienii: Keep forks synchronized.`,
		Config:                &types.NoOption{},
		DefaultPointersConfig: &types.NoOption{},
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

	generateCmd := &flaeg.Command{
		DefaultPointersConfig: &types.GenerateOptions{},
		Description:           "Generate configuration file.",
		Name:                  "gen",
		Config:                &types.GenerateOptions{},
		Run:                   runGenerate(&types.GenerateOptions{}),
	}

	flag.AddCommand(generateCmd)

	// version

	versionCmd := &flaeg.Command{
		Name:                  "version",
		Description:           "Display the version.",
		Config:                &types.NoOption{},
		DefaultPointersConfig: &types.NoOption{},
		Run: func() error {
			displayVersion()
			return nil
		},
	}

	flag.AddCommand(versionCmd)

	// Print help when the command is running without any parameters.
	rootCmd.Run = func() error {
		return flaeg.LoadWithCommand(rootCmd, []string{"-h"}, nil, []*flaeg.Command{rootCmd, syncCmd, generateCmd, versionCmd})
	}

	// Run command
	err := flag.Run()
	if err != nil && !errors.Is(err, pflag.ErrHelp) {
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

// NewGitHubClient create a new GitHub client.
func NewGitHubClient(ctx context.Context, token string) *github.Client {
	if len(token) == 0 {
		return github.NewClient(nil)
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
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

func required(field, fieldName string) error {
	if len(field) == 0 {
		return fmt.Errorf("%s is mandatory", fieldName)
	}
	return nil
}
