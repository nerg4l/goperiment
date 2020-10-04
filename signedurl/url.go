package signedurl

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

func Sign(ref *url.URL, key []byte) (*url.URL, error) {
	return sign(ref, time.Time{}, key)
}

func SignWithExpiration(ref *url.URL, expires time.Time, key []byte) (*url.URL, error) {
	return sign(ref, expires, key)
}

func sign(ref *url.URL, expires time.Time, key []byte) (*url.URL, error) {
	u := *ref
	if !u.IsAbs() {
		return nil, fmt.Errorf("signedurl: only absolute urls can be signed")
	}
	q := u.Query()
	if _, ok := q["signature"]; ok {
		return nil, fmt.Errorf("signedurl: signature is a preserved query parameter")
	}
	if !expires.IsZero() {
		q["expires"] = []string{strconv.FormatInt(expires.Unix(), 10)}
	}
	u.RawQuery = q.Encode()
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(u.String()))
	signature := hex.EncodeToString(mac.Sum(nil))
	q.Add("signature", signature)
	u.RawQuery = q.Encode()
	return &u, nil
}

func Middleware(baseURL *url.URL, key []byte) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			query := r.URL.Query()
			if _, ok := query["signature"]; !ok {
				http.Error(w, "403 Forbidden", http.StatusForbidden)
				return
			}
			ok, err := ValidSignature(baseURL.ResolveReference(r.URL), key)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			if !ok {
				http.Error(w, "403 Forbidden", http.StatusForbidden)
				return
			}
			if Expired(r.URL) {
				http.Error(w, "403 Forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func ValidSignature(ref *url.URL, key []byte) (bool, error) {
	u := *ref
	if !u.IsAbs() {
		return false, fmt.Errorf("signedurl: only absolute urls can be signed")
	}
	query := u.Query()
	signature := query.Get("signature")
	query.Del("signature")
	u.RawQuery = query.Encode()
	s, err := hex.DecodeString(signature)
	if err != nil {
		return false, nil
	}
	return validMAC([]byte(u.String()), s, key), nil
}

func validMAC(message, messageMAC, key []byte) bool {
	mac := hmac.New(sha256.New, key)
	mac.Write(message)
	expectedMAC := mac.Sum(nil)
	return hmac.Equal(messageMAC, expectedMAC)
}

func Expired(u *url.URL) bool {
	expires := u.Query().Get("expires")
	if len(expires) == 0 {
		return false
	}
	i, err := strconv.ParseInt(expires, 10, 64)
	if err != nil {
		return true
	}
	tm := time.Unix(i, 0)
	return tm.Before(time.Now())
}
