package utils

import (
	"net/http"
	"time"
)

func newCookie(name, value string) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   int(time.Now().AddDate(0, 0, 7).Unix()),
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
	}
}

func read(r *http.Request, key string) string {
	token, err := r.Cookie(key)
	if err != nil {
		return ""
	}
	return token.Value
}

func SetJWTCookie(w http.ResponseWriter, token string) {
	c := newCookie("token", token)
	http.SetCookie(w, c)
}

func ReadJWTCookie(r *http.Request) string {
	return read(r, "token")
}

func DeleteJWTCookie(w http.ResponseWriter) {
	c := newCookie("token", "")
	c.Expires = time.Unix(0, 0)
	http.SetCookie(w, c)
}
