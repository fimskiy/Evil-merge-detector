package ghclient

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bradleyfalzon/ghinstallation/v2"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
)

// Clone clones repoURL into destDir using a fresh installation token.
// branch specifies which branch to clone; full clone is required for merge-base analysis.
func Clone(ctx context.Context, appID, installationID int64, privateKey []byte, repoURL, branch, destDir string) error {
	tr, err := ghinstallation.New(http.DefaultTransport, appID, installationID, privateKey)
	if err != nil {
		return fmt.Errorf("creating transport: %w", err)
	}

	token, err := tr.Token(ctx)
	if err != nil {
		return fmt.Errorf("getting installation token: %w", err)
	}

	opts := &git.CloneOptions{
		URL: repoURL,
		Auth: &githttp.BasicAuth{
			Username: "x-access-token",
			Password: token,
		},
	}
	if branch != "" {
		opts.ReferenceName = plumbing.NewBranchReferenceName(branch)
	}

	_, err = git.PlainCloneContext(ctx, destDir, false, opts)
	if err != nil {
		return fmt.Errorf("cloning %s: %w", repoURL, err)
	}
	return nil
}
