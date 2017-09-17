package sync

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"text/template"

	"github.com/containous/gallienii/types"
	"github.com/google/go-github/github"
)

type message struct {
	title string
	body  string
}

const bodyTemplate = `
The repository [{{ .Owner }}/{{ .Name }}](https://github.com/{{ .Owner }}/{{ .Name }}/tree/{{ .Branch}}) has some new changes that aren't in this fork.

:robot::speech_balloon: _Done with :heart: by :ant: [Myrmica Gallienii](https://github.com/containous/gallienii) :ant:_
`

// Process synchronize forks by making Pull Request.
func Process(ctx context.Context, client *github.Client, configs *types.Configuration, dryRun bool, verbose bool) error {
	for _, conf := range configs.Forks {
		if !conf.Disable {
			err := processOneRepository(ctx, client, conf, dryRun, verbose)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func processOneRepository(ctx context.Context, client *github.Client, forkConfig types.ForkConfiguration, dryRun bool, verbose bool) error {

	if !forkConfig.NoCheckFork {
		err := checkFork(ctx, client, forkConfig.Fork)
		if err != nil {
			return err
		}
	}

	forkRef := fmt.Sprintf("%s:%s", forkConfig.Fork.Owner, forkConfig.Fork.Branch)

	cc, _, err := client.Repositories.CompareCommits(ctx, forkConfig.Base.Owner, forkConfig.Base.Name, forkConfig.Base.Branch, forkRef)
	if err != nil {
		return err
	}

	if verbose {
		log.Println("MergeBaseCommit", cc.MergeBaseCommit.GetSHA())
		log.Println("Status", cc.GetStatus())
		log.Println("TotalCommits", cc.GetTotalCommits())
		log.Println("AheadBy", cc.GetAheadBy())
		log.Println("BehindBy", cc.GetBehindBy())
	}

	if cc.GetBehindBy() > 0 {
		msg := makeMessage(forkConfig.Base)

		pr, errPr := createPullRequest(ctx, client, forkConfig.Fork, forkConfig.Base, msg, dryRun)
		if errPr != nil {
			return errPr
		}

		if pr != nil {
			log.Printf("PR done: %s", pr.GetHTMLURL())

			errLabel := addLabels(ctx, client, pr, forkConfig.Marker)
			if errLabel != nil {
				return errLabel
			}
		}
	}
	return nil
}

func checkFork(ctx context.Context, client *github.Client, fork types.Repo) error {
	repo, _, err := client.Repositories.Get(ctx, fork.Owner, fork.Name)
	if err != nil {
		return err
	}

	if !repo.GetFork() {
		return fmt.Errorf("%s is not a fork", fork)
	}

	return nil
}

func addLabels(ctx context.Context, client *github.Client, pr *github.PullRequest, marker types.Marker) error {
	labels := []string{}

	if !pr.GetMergeable() && marker.NeedResolveConflicts != "" {
		labels = append(labels, marker.NeedResolveConflicts)
	}

	if marker.ByBot != "" {
		labels = append(labels, marker.ByBot)
	}

	_, _, err := client.Issues.AddLabelsToIssue(ctx, pr.Base.Repo.Owner.GetLogin(), pr.Base.Repo.GetName(), pr.GetNumber(), labels)

	return err
}

func makeMessage(base types.Repo) message {
	tmpl := template.Must(template.New("pr").Parse(bodyTemplate))

	b := &bytes.Buffer{}
	err := tmpl.Execute(b, base)
	if err != nil {
		log.Fatal(err)
	}

	return message{
		title: fmt.Sprintf("🤖 Update from upstream repository %s/%s", base.Owner, base.Name),
		body:  b.String(),
	}
}

func createPullRequest(ctx context.Context, client *github.Client, fork types.Repo, base types.Repo, msg message, dryRun bool) (*github.PullRequest, error) {
	newPR := &github.NewPullRequest{
		Title:               github.String(msg.title),
		Head:                github.String(base.Owner + ":" + base.Branch),
		Base:                github.String(fork.Branch),
		Body:                github.String(msg.body),
		MaintainerCanModify: github.Bool(true),
	}

	if dryRun {
		log.Printf("PR: Head %s, Base %s", newPR.GetHead(), newPR.GetBase())
		log.Printf("Title: %s", newPR.GetTitle())
		log.Printf("Body: %s", newPR.GetBody())
		return nil, nil
	}

	pr, _, err := client.PullRequests.Create(ctx, fork.Owner, fork.Name, newPR)
	if err != nil {
		return nil, fmt.Errorf("unable to create the PR: %v", err)
	}

	return pr, nil
}
