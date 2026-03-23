package ghclient

import (
	"net/http"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v84/github"
)

// ForInstallation returns a GitHub client authenticated as the given installation.
func ForInstallation(appID, installationID int64, privateKey []byte) (*github.Client, error) {
	tr, err := ghinstallation.New(http.DefaultTransport, appID, installationID, privateKey)
	if err != nil {
		return nil, err
	}
	return github.NewClient(&http.Client{Transport: tr}), nil
}
