package auth

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrTokenExpired = errors.New("token expired")
)

type Claims struct {
	Subject  string
	ClientID string
	Scope    string
}

type Verifier struct {
	jwksURL    string
	httpClient *http.Client
	mu         sync.RWMutex
	keys       map[string]*ecdsa.PublicKey
	lastFetch  time.Time
	refreshTTL time.Duration
}

func NewVerifier(jwksURL string) *Verifier {
	return &Verifier{
		jwksURL: jwksURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		keys:       make(map[string]*ecdsa.PublicKey),
		refreshTTL: 5 * time.Minute,
	}
}

type jwksDocument struct {
	Keys []jwkKey `json:"keys"`
}

type jwkKey struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Crv string `json:"crv"`
	X   string `json:"x"`
	Y   string `json:"y"`
}

func (v *Verifier) Verify(ctx context.Context, tokenString string) (*Claims, error) {
	if err := v.ensureKeys(ctx, ""); err != nil {
		return nil, err
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if token.Method.Alg() != jwt.SigningMethodES256.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %s", token.Method.Alg())
		}

		kid, _ := token.Header["kid"].(string)
		if kid == "" {
			return nil, ErrInvalidToken
		}

		key, ok := v.getKey(kid)
		if !ok {
			if refreshErr := v.ensureKeys(ctx, kid); refreshErr != nil {
				return nil, refreshErr
			}
			key, ok = v.getKey(kid)
			if !ok {
				return nil, ErrInvalidToken
			}
		}
		return key, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodES256.Alg()}))
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}
	if !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	sub, _ := claims["sub"].(string)
	if sub == "" {
		return nil, ErrInvalidToken
	}

	clientID, _ := claims["azp"].(string)
	if clientID == "" {
		return nil, ErrInvalidToken
	}

	scope, _ := claims["scope"].(string)

	return &Claims{
		Subject:  sub,
		ClientID: clientID,
		Scope:    scope,
	}, nil
}

func (v *Verifier) getKey(kid string) (*ecdsa.PublicKey, bool) {
	v.mu.RLock()
	defer v.mu.RUnlock()
	key, ok := v.keys[kid]
	return key, ok
}

func (v *Verifier) ensureKeys(ctx context.Context, requiredKid string) error {
	v.mu.RLock()
	needsRefresh := len(v.keys) == 0 || time.Since(v.lastFetch) > v.refreshTTL
	if !needsRefresh && requiredKid != "" {
		_, ok := v.keys[requiredKid]
		needsRefresh = !ok
	}
	v.mu.RUnlock()

	if !needsRefresh {
		return nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, v.jwksURL, nil)
	if err != nil {
		return fmt.Errorf("build jwks request: %w", err)
	}

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("fetch jwks: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fetch jwks: status=%d", resp.StatusCode)
	}

	var doc jwksDocument
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return fmt.Errorf("decode jwks: %w", err)
	}

	keys := make(map[string]*ecdsa.PublicKey, len(doc.Keys))
	for _, key := range doc.Keys {
		if key.Kty != "EC" || key.Crv != "P-256" {
			continue
		}
		pub, err := parseECPublicKey(key.X, key.Y)
		if err != nil {
			continue
		}
		keys[key.Kid] = pub
	}

	v.mu.Lock()
	v.keys = keys
	v.lastFetch = time.Now()
	v.mu.Unlock()

	return nil
}

func parseECPublicKey(xRaw, yRaw string) (*ecdsa.PublicKey, error) {
	xBytes, err := base64.RawURLEncoding.DecodeString(xRaw)
	if err != nil {
		return nil, err
	}
	yBytes, err := base64.RawURLEncoding.DecodeString(yRaw)
	if err != nil {
		return nil, err
	}

	return &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     new(big.Int).SetBytes(xBytes),
		Y:     new(big.Int).SetBytes(yBytes),
	}, nil
}
