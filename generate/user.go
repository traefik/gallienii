package generate

import (
	"context"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/containous/gallienii/types"
	"github.com/google/go-github/v27/github"
)

// UserConfiguration generate a default configuration file for a user.
func UserConfiguration(ctx context.Context, client *github.Client, user, path string) error {
	opt := &github.RepositoryListOptions{
		Type: "forks",
	}

	repos, _, err := client.Repositories.List(ctx, user, opt)
	if err != nil {
		return err
	}

	var configs []types.ForkConfiguration

	for _, rep := range repos {
		repo, _, errRepo := client.Repositories.Get(ctx, rep.Owner.GetLogin(), rep.GetName())
		if errRepo != nil {
			return errRepo
		}

		conf := types.ForkConfiguration{
			Fork: types.Repo{
				Branch: repo.GetDefaultBranch(),
				Owner:  repo.Owner.GetLogin(),
				Name:   repo.GetName(),
			},
			Base: types.Repo{
				Branch: repo.Source.GetDefaultBranch(),
				Owner:  repo.Source.Owner.GetLogin(),
				Name:   repo.Source.GetName(),
			},
		}
		configs = append(configs, conf)
	}

	cnf := types.Configuration{Forks: configs}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	return toml.NewEncoder(f).Encode(cnf)
}
