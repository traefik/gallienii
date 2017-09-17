package generate

import (
	"os"

	"github.com/BurntSushi/toml"
	"github.com/containous/gallienii/types"
)

// Sample generate a sample configuration file.
func Sample(path string) error {
	configs := types.Configuration{
		Forks: []types.ForkConfiguration{
			{
				Base: types.Repo{Name: "linux", Branch: "master", Owner: "torvalds"},
				Fork: types.Repo{Name: "linux", Branch: "master", Owner: "login"},
				Marker: types.Marker{
					NeedResolveConflicts: "human/need-resolve-conflicts",
					ByBot:                "bot/upstream-sync",
				},
			},
			{
				Base: types.Repo{Name: "moby", Branch: "master", Owner: "moby"},
				Fork: types.Repo{Name: "moby", Branch: "master", Owner: "login"},
				Marker: types.Marker{
					NeedResolveConflicts: "human/need-resolve-conflicts",
					ByBot:                "bot/upstream-sync",
				},
				NoCheckFork: true,
			},
			{
				Base: types.Repo{Name: "kubernetes", Branch: "master", Owner: "kubernetes"},
				Fork: types.Repo{Name: "kubernetes", Branch: "master", Owner: "login"},
				Marker: types.Marker{
					NeedResolveConflicts: "human/need-resolve-conflicts",
				},
			},
			{
				Base: types.Repo{Name: "go", Branch: "master", Owner: "golang"},
				Fork: types.Repo{Name: "go", Branch: "master", Owner: "login"},
			},
		},
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}

	return toml.NewEncoder(f).Encode(configs)
}
