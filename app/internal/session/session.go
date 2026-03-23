package session

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

const cookieName = "emd_session"

type Data struct {
	GitHubLogin    string `json:"login"`
	GitHubID       int64  `json:"id"`
	InstallationID int64  `json:"iid"`
}

func Get(r *http.Request, secret []byte) (*Data, bool) {
	c, err := r.Cookie(cookieName)
	if err != nil {
		return nil, false
	}
	data, ok := verify(c.Value, secret)
	return data, ok
}

func Set(w http.ResponseWriter, d *Data, secret []byte) {
	val := sign(d, secret)
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    val,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(30 * 24 * time.Hour),
	})
}

func Clear(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:    cookieName,
		Value:   "",
		Path:    "/",
		Expires: time.Unix(0, 0),
	})
}

func sign(d *Data, secret []byte) string {
	payload, _ := json.Marshal(d)
	enc := base64.RawURLEncoding.EncodeToString(payload)
	mac := mac(enc, secret)
	return enc + "." + mac
}

func verify(val string, secret []byte) (*Data, bool) {
	parts := strings.SplitN(val, ".", 2)
	if len(parts) != 2 {
		return nil, false
	}
	if !hmac.Equal([]byte(parts[1]), []byte(mac(parts[0], secret))) {
		return nil, false
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, false
	}
	var d Data
	if err := json.Unmarshal(payload, &d); err != nil {
		return nil, false
	}
	return &d, true
}

func mac(data string, secret []byte) string {
	h := hmac.New(sha256.New, secret)
	h.Write([]byte(data))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}
