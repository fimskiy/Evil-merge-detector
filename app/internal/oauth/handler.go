package oauth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"

	"github.com/fimskiy/evil-merge-detector/app/internal/session"
	"github.com/fimskiy/evil-merge-detector/app/internal/store"
)

type Handler struct {
	cfg    *oauth2.Config
	secret []byte
	db     *store.Store
}

func New(clientID, clientSecret string, secret []byte, db *store.Store) *Handler {
	return &Handler{
		cfg: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Scopes:       []string{"read:user"},
			Endpoint:     github.Endpoint,
		},
		secret: secret,
		db:     db,
	}
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	state, err := randomState()
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   300,
	})
	http.Redirect(w, r, h.cfg.AuthCodeURL(state), http.StatusFound)
}

func (h *Handler) Callback(w http.ResponseWriter, r *http.Request) {
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil || stateCookie.Value != r.FormValue("state") {
		http.Error(w, "invalid state", http.StatusBadRequest)
		return
	}
	http.SetCookie(w, &http.Cookie{Name: "oauth_state", MaxAge: -1, Path: "/"})

	token, err := h.cfg.Exchange(r.Context(), r.FormValue("code"))
	if err != nil {
		http.Error(w, "token exchange failed", http.StatusInternalServerError)
		return
	}

	user, err := fetchUser(r.Context(), h.cfg, token)
	if err != nil {
		http.Error(w, "failed to fetch user", http.StatusInternalServerError)
		return
	}

	var installationID int64
	if h.db != nil {
		inst, err := h.db.GetInstallationByLogin(r.Context(), user.Login)
		if err == nil {
			installationID = inst.InstallationID
		}
	}

	session.Set(w, &session.Data{
		GitHubLogin:    user.Login,
		GitHubID:       user.ID,
		InstallationID: installationID,
	}, h.secret)

	http.Redirect(w, r, "/dashboard", http.StatusFound)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	session.Clear(w)
	http.Redirect(w, r, "/", http.StatusFound)
}

type ghUser struct {
	Login string `json:"login"`
	ID    int64  `json:"id"`
}

func fetchUser(ctx context.Context, cfg *oauth2.Config, token *oauth2.Token) (*ghUser, error) {
	client := cfg.Client(ctx, token)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github api: %s", resp.Status)
	}
	var u ghUser
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return nil, err
	}
	return &u, nil
}

func randomState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		log.Printf("oauth: rand.Read: %v", err)
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
