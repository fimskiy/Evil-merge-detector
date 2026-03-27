package ghclient

import (
	"context"

	"github.com/google/go-github/v84/github"
)

type Repo struct {
	Owner         string
	Name          string
	CloneURL      string
	DefaultBranch string
}

// ListRepos returns all repositories accessible to the given installation.
func ListRepos(ctx context.Context, appID, installationID int64, privateKey []byte) ([]Repo, error) {
	client, err := ForInstallation(appID, installationID, privateKey)
	if err != nil {
		return nil, err
	}

	var repos []Repo
	opts := &github.ListOptions{PerPage: 100}
	for {
		result, resp, err := client.Apps.ListRepos(ctx, opts)
		if err != nil {
			return nil, err
		}
		for _, r := range result.Repositories {
			repos = append(repos, Repo{
				Owner:         r.GetOwner().GetLogin(),
				Name:          r.GetName(),
				CloneURL:      r.GetCloneURL(),
				DefaultBranch: r.GetDefaultBranch(),
			})
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return repos, nil
}
