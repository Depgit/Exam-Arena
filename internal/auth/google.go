package auth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// googleCertsURL publishes the public keys Google signs its sign-in tokens with.
const googleCertsURL = "https://www.googleapis.com/oauth2/v3/certs"

// GoogleIdentity is what a verified "Sign in with Google" token tells us.
type GoogleIdentity struct {
	Sub     string // Google's stable account id
	Email   string
	Name    string
	Picture string
}

var errBadGoogleToken = errors.New("Google sign-in failed — please try again")

// googleVerifier checks the ID token the "Sign in with Google" button hands
// the browser: signed by Google, meant for our client id, not expired, and
// for an email Google has verified. No client secret is involved.
type googleVerifier struct {
	clientID string
	certsURL string
	client   *http.Client

	mu        sync.Mutex
	keys      map[string]*rsa.PublicKey
	expires   time.Time // when Google says the cached keys go stale
	lastFetch time.Time
}

func newGoogleVerifier(clientID string) *googleVerifier {
	return &googleVerifier{
		clientID: clientID,
		certsURL: googleCertsURL,
		client:   &http.Client{Timeout: 10 * time.Second},
	}
}

type googleClaims struct {
	Email         string      `json:"email"`
	EmailVerified interface{} `json:"email_verified"` // bool, sometimes the string "true"
	Name          string      `json:"name"`
	Picture       string      `json:"picture"`
	jwt.RegisteredClaims
}

func (v *googleVerifier) Verify(ctx context.Context, idToken string) (*GoogleIdentity, error) {
	if v.clientID == "" {
		return nil, errors.New("Google sign-in is not set up on this server")
	}
	claims := &googleClaims{}
	_, err := jwt.ParseWithClaims(idToken, claims,
		func(t *jwt.Token) (interface{}, error) {
			kid, _ := t.Header["kid"].(string)
			return v.key(ctx, kid)
		},
		jwt.WithValidMethods([]string{"RS256"}),
		jwt.WithAudience(v.clientID),
		jwt.WithExpirationRequired(),
		jwt.WithLeeway(30*time.Second),
	)
	if err != nil {
		return nil, errBadGoogleToken
	}
	if iss := claims.Issuer; iss != "accounts.google.com" && iss != "https://accounts.google.com" {
		return nil, errBadGoogleToken
	}
	verified := claims.EmailVerified == true || claims.EmailVerified == "true"
	if claims.Subject == "" || claims.Email == "" || !verified {
		return nil, errors.New("your Google account's email isn't verified")
	}
	return &GoogleIdentity{
		Sub:     claims.Subject,
		Email:   strings.ToLower(claims.Email),
		Name:    claims.Name,
		Picture: claims.Picture,
	}, nil
}

// key returns Google's public key with this id, refreshing the cached set
// when it is stale or the id is new (Google rotates keys).
func (v *googleVerifier) key(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if k, ok := v.keys[kid]; ok && time.Now().Before(v.expires) {
		return k, nil
	}
	// An unknown kid could be junk; refetch at most every 30 seconds.
	if time.Since(v.lastFetch) > 30*time.Second || v.keys == nil {
		if err := v.fetch(ctx); err != nil {
			return nil, err
		}
	}
	if k, ok := v.keys[kid]; ok {
		return k, nil
	}
	return nil, fmt.Errorf("unknown Google key %q", kid)
}

func (v *googleVerifier) fetch(ctx context.Context) error {
	v.lastFetch = time.Now()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, v.certsURL, nil)
	if err != nil {
		return err
	}
	resp, err := v.client.Do(req)
	if err != nil {
		return fmt.Errorf("fetch Google keys: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fetch Google keys: status %d", resp.StatusCode)
	}
	var set struct {
		Keys []struct {
			Kid string `json:"kid"`
			Kty string `json:"kty"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&set); err != nil {
		return fmt.Errorf("decode Google keys: %w", err)
	}
	keys := map[string]*rsa.PublicKey{}
	for _, k := range set.Keys {
		if k.Kty != "RSA" {
			continue
		}
		n, err1 := base64.RawURLEncoding.DecodeString(k.N)
		e, err2 := base64.RawURLEncoding.DecodeString(k.E)
		if err1 != nil || err2 != nil {
			continue
		}
		keys[k.Kid] = &rsa.PublicKey{N: new(big.Int).SetBytes(n), E: int(new(big.Int).SetBytes(e).Int64())}
	}
	if len(keys) == 0 {
		return errors.New("Google returned no usable keys")
	}
	v.keys = keys
	v.expires = time.Now().Add(maxAge(resp.Header.Get("Cache-Control"), time.Hour))
	return nil
}

// maxAge reads "max-age=N" from a Cache-Control header.
func maxAge(cacheControl string, fallback time.Duration) time.Duration {
	for _, part := range strings.Split(cacheControl, ",") {
		part = strings.TrimSpace(part)
		if s, ok := strings.CutPrefix(part, "max-age="); ok {
			if n, err := strconv.Atoi(s); err == nil && n > 0 {
				return time.Duration(n) * time.Second
			}
		}
	}
	return fallback
}
