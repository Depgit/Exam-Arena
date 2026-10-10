package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testClientID = "test-client.apps.googleusercontent.com"

// fakeGoogle serves a key set like Google's and signs tokens with it.
func fakeGoogle(t *testing.T) (*googleVerifier, func(claims jwt.MapClaims) string) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=3600")
		json.NewEncoder(w).Encode(map[string]interface{}{"keys": []map[string]string{{
			"kid": "k1", "kty": "RSA", "alg": "RS256",
			"n": base64.RawURLEncoding.EncodeToString(key.N.Bytes()),
			"e": base64.RawURLEncoding.EncodeToString(big.NewInt(int64(key.E)).Bytes()),
		}}})
	}))
	t.Cleanup(srv.Close)
	v := newGoogleVerifier(testClientID)
	v.certsURL = srv.URL
	sign := func(claims jwt.MapClaims) string {
		tok := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		tok.Header["kid"] = "k1"
		s, err := tok.SignedString(key)
		if err != nil {
			t.Fatal(err)
		}
		return s
	}
	return v, sign
}

func goodClaims() jwt.MapClaims {
	return jwt.MapClaims{
		"iss": "https://accounts.google.com", "aud": testClientID, "sub": "1234567890",
		"email": "Priya@Gmail.com", "email_verified": true, "name": "Priya",
		"exp": time.Now().Add(time.Hour).Unix(), "iat": time.Now().Unix(),
	}
}

func TestGoogleVerifierAcceptsAGoodToken(t *testing.T) {
	v, sign := fakeGoogle(t)
	id, err := v.Verify(context.Background(), sign(goodClaims()))
	if err != nil {
		t.Fatalf("good token rejected: %v", err)
	}
	if id.Sub != "1234567890" || id.Email != "priya@gmail.com" || id.Name != "Priya" {
		t.Fatalf("unexpected identity %+v", id)
	}
}

func TestGoogleVerifierRejectsBadTokens(t *testing.T) {
	v, sign := fakeGoogle(t)
	cases := map[string]func(jwt.MapClaims){
		"another app's token": func(c jwt.MapClaims) { c["aud"] = "someone-else" },
		"expired":             func(c jwt.MapClaims) { c["exp"] = time.Now().Add(-time.Hour).Unix() },
		"not from Google":     func(c jwt.MapClaims) { c["iss"] = "https://evil.example" },
		"unverified email":    func(c jwt.MapClaims) { c["email_verified"] = false },
		"no email":            func(c jwt.MapClaims) { delete(c, "email") },
	}
	for name, mutate := range cases {
		c := goodClaims()
		mutate(c)
		if _, err := v.Verify(context.Background(), sign(c)); err == nil {
			t.Errorf("%s: accepted, want rejected", name)
		}
	}
	// Signed by a key Google never published.
	other, _ := rsa.GenerateKey(rand.Reader, 2048)
	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, goodClaims())
	tok.Header["kid"] = "k1"
	forged, _ := tok.SignedString(other)
	if _, err := v.Verify(context.Background(), forged); err == nil {
		t.Error("forged signature: accepted, want rejected")
	}
}

func TestGoogleOffWithoutClientID(t *testing.T) {
	if _, err := newGoogleVerifier("").Verify(context.Background(), "x"); err == nil {
		t.Fatal("want an error when GOOGLE_CLIENT_ID is not set")
	}
}

func TestValidateEmail(t *testing.T) {
	for email, ok := range map[string]bool{
		"priya@gmail.com":     true,
		"a.b@college.ac.in":   true,
		"nope":                false,
		"x@localhost":         false,
		"bot@mailinator.com":  false,
		"two words@gmail.com": false,
		"someone@yopmail.com": false,
	} {
		if err := validateEmail(email); (err == nil) != ok {
			t.Errorf("%q: err=%v, want ok=%v", email, err, ok)
		}
	}
}

func TestVerificationCodeShapeAndHash(t *testing.T) {
	code, err := newCode()
	if err != nil || len(code) != 6 {
		t.Fatalf("code %q, err %v", code, err)
	}
	if hashCode("u1", code) == hashCode("u2", code) {
		t.Fatal("the same code must hash differently for different players")
	}
}
