package session

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

var testSecret = []byte("test-secret-32-bytes-long-enough!")

func TestSetGet_roundtrip(t *testing.T) {
	d := &Data{GitHubLogin: "alice", GitHubID: 42, InstallationID: 7}

	w := httptest.NewRecorder()
	Set(w, d, testSecret)

	r := &http.Request{Header: http.Header{"Cookie": w.Result().Header["Set-Cookie"]}}
	got, ok := Get(r, testSecret)
	if !ok {
		t.Fatal("Get returned false")
	}
	if got.GitHubLogin != d.GitHubLogin || got.GitHubID != d.GitHubID || got.InstallationID != d.InstallationID {
		t.Fatalf("got %+v, want %+v", got, d)
	}
}

func TestGet_wrongSecret(t *testing.T) {
	w := httptest.NewRecorder()
	Set(w, &Data{GitHubLogin: "alice"}, testSecret)

	r := &http.Request{Header: http.Header{"Cookie": w.Result().Header["Set-Cookie"]}}
	_, ok := Get(r, []byte("different-secret-value-here-!!!"))
	if ok {
		t.Fatal("expected false with wrong secret")
	}
}

func TestGet_noCookie(t *testing.T) {
	r := &http.Request{Header: http.Header{}}
	_, ok := Get(r, testSecret)
	if ok {
		t.Fatal("expected false with no cookie")
	}
}

func TestGet_tamperedPayload(t *testing.T) {
	// Directly set a cookie with a modified value (valid base64.sig format but wrong HMAC)
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.AddCookie(&http.Cookie{Name: cookieName, Value: "dGVzdA.invalidsignatureXXXXXXXXXXXXXXXXXXXXXXXXXXXX"})
	_, ok := Get(r, testSecret)
	if ok {
		t.Fatal("expected false with invalid HMAC")
	}
}

func TestGet_truncatedValue(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.AddCookie(&http.Cookie{Name: cookieName, Value: "nodot"})
	_, ok := Get(r, testSecret)
	if ok {
		t.Fatal("expected false with truncated value")
	}
}

func TestClear_removesSession(t *testing.T) {
	w := httptest.NewRecorder()
	Set(w, &Data{GitHubLogin: "alice"}, testSecret)

	wClear := httptest.NewRecorder()
	Clear(wClear)

	c := wClear.Result().Cookies()
	if len(c) == 0 {
		t.Fatal("expected a cookie in Clear response")
	}
	if c[0].MaxAge >= 0 && c[0].Expires.Unix() > 1 {
		t.Fatal("expected cookie to be expired")
	}
}
