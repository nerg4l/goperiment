package signedurl

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func urlParse(rawurl string) *url.URL {
	u, err := url.Parse(rawurl)
	if err != nil {
		panic(err)
	}
	return u
}

func TestSign(t *testing.T) {
	secret := []byte("secret")
	u, err := Sign(urlParse("https://example.com/1"), secret)
	if err != nil {
		t.Fatal(err)
	}
	ok, err := ValidSignature(u, secret)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Error("Signed url must be valid")
	}
	ok = Expired(u)
	if ok {
		t.Error("Signed url must be valid")
	}
}

func TestSignWithExpiration(t *testing.T) {
	secret := []byte("secret")
	u, err := SignWithExpiration(
		urlParse("https://example.com/1"),
		time.Now().Add(time.Second),
		secret,
	)
	if err != nil {
		t.Fatal(err)
	}
	ok, err := ValidSignature(u, secret)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Error("Signed url must be valid")
	}
	ok = Expired(u)
	if ok {
		t.Error("Signed url must be valid")
	}

	u, err = SignWithExpiration(
		urlParse("https://example.com/1"),
		time.Now().Add(-time.Second),
		secret,
	)
	if err != nil {
		t.Fatal(err)
	}
	ok = Expired(u)
	if !ok {
		t.Error("Signed url must be invalid")
	}
}

func TestMiddleware(t *testing.T) {
	sign := func(rawurl string, t time.Time, key []byte) string {
		sign, _ := sign(urlParse(rawurl), t, key)
		return sign.String()
	}
	type args struct {
		baseURL *url.URL
		key     []byte
	}
	tests := []struct {
		name     string
		args     args
		target   string
		wantCode int
		wantBody string
	}{
		{
			name: "missing param",
			args: args{
				baseURL: urlParse("http://example.com/"),
				key:     []byte("0001020304050607"),
			},
			target:   "http://example.com/",
			wantCode: 403,
			wantBody: "403 Forbidden",
		},
		{
			name: "invalid param",
			args: args{
				baseURL: urlParse("http://example.com/"),
				key:     []byte("0001020304050607"),
			},
			target:   "http://example.com/?signature=invalid",
			wantCode: 403,
			wantBody: "403 Forbidden",
		},
		{
			name: "expired",
			args: args{
				baseURL: urlParse("http://example.com/"),
				key:     []byte("0001020304050607"),
			},
			target:   sign("http://example.com/", time.Now().Add(-time.Second), []byte("0001020304050607")),
			wantCode: 403,
			wantBody: "403 Forbidden",
		},
		{
			name: "valid",
			args: args{
				baseURL: urlParse("http://example.com/"),
				key:     []byte("0001020304050607"),
			},
			target:   sign("http://example.com/", time.Time{}, []byte("0001020304050607")),
			wantCode: 200,
			wantBody: "success",
		},
		{
			name: "valid with expiration",
			args: args{
				baseURL: urlParse("http://example.com/"),
				key:     []byte("0001020304050607"),
			},
			target:   sign("http://example.com/", time.Now().Add(time.Second), []byte("0001020304050607")),
			wantCode: 200,
			wantBody: "success",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := Middleware(tt.args.baseURL, tt.args.key)
			handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				fmt.Fprintln(w, "success")
			})

			signed := handler(handlerFunc)
			recorder := httptest.NewRecorder()
			signed.ServeHTTP(recorder, httptest.NewRequest("GET", tt.target, nil))

			if gotCode := recorder.Code; gotCode != tt.wantCode {
				t.Fatalf("ServeHTTP() gotCode = %v, wantCode = %v", gotCode, tt.wantCode)
			}
			if gotBody := recorder.Body.String(); !strings.Contains(gotBody, tt.wantBody) {
				t.Fatalf("ServeHTTP() gotBody = %v, wantBody = %v", gotBody, tt.wantBody)
			}
		})
	}
}

func TestValidSignature(t *testing.T) {
	type args struct {
		ref *url.URL
		key []byte
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "relative url",
			args: args{
				ref: urlParse("/feed/news/11?signature=fb5b9ae9f476eb6dfd7cb7db4df3b84e262922c7636c5b441edf4c8b8ffa4977"),
				key: []byte("0001020304050607"),
			},
			wantErr: true,
		},
		{
			name: "invalid",
			args: args{
				ref: urlParse("https://example.com/1?signature=invalid"),
				key: []byte("0001020304050607"),
			},
			want: false,
		},
		{
			name: "valid",
			args: args{
				ref: urlParse("https://example.com/1?signature=62540c19a073672e1a88c9353fb97b5d94bc52a4bee52e58dbf171036e05fe4b"),
				key: []byte("0001020304050607"),
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidSignature(tt.args.ref, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidSignature() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ValidSignature() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExpired(t *testing.T) {
	type args struct {
		u *url.URL
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "empty",
			args: args{
				u: urlParse("https://example.com/1"),
			},
			want: false,
		},
		{
			name: "invalid",
			args: args{
				u: urlParse("https://example.com/1?expires=asd"),
			},
			want: true,
		},
		{
			name: "expired",
			args: args{
				u: urlParse(fmt.Sprintf("https://example.com/1?expires=%d", time.Now().Add(-time.Second).Unix())),
			},
			want: true,
		},
		{
			name: "not expired",
			args: args{
				u: urlParse(fmt.Sprintf("https://example.com/1?expires=%d", time.Now().Add(time.Second).Unix())),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Expired(tt.args.u); got != tt.want {
				t.Errorf("Expired() = %v, want %v", got, tt.want)
			}
		})
	}
}
