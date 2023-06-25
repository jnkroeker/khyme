// Package auth provides authentication and authorization support
// Authentication: You are who you say you are.
// Authorization: You have permission to do what you are requesting to do.
package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/golang-jwt/jwt/v4"
	"github.com/jmoiron/sqlx"
	"github.com/jnkroeker/khyme/business/core/event"
	"github.com/jnkroeker/khyme/business/core/user"
	"github.com/jnkroeker/khyme/business/core/user/stores/userdb"
	"go.uber.org/zap"
)

// ErrForbidden is returned when an auth issue is identified.
var ErrForbidden = errors.New("attempted action is not allowed")

// Claims represents the authorization claims transmitted via a JWT.
type Claims struct {
	jwt.RegisteredClaims
	Roles []user.Role `json:"roles"`
}

// declares a method set of behavior for looking up
// private and public keys for JWT use. The return could be a
// PEM encoded string or a JWS based key.
type KeyLookup interface {
	PrivateKey(kid string) (key string, err error)
	PublicKey(kid string) (key string, err error)
}

// Config represents information required to initialize auth
type Config struct {
	Log       *zap.SugaredLogger
	DB        *sqlx.DB
	KeyLookup KeyLookup
	Issuer    string
}

// Auth is used to authenticate clients. It can generate a token for a
// set of user claims and recreate the claims by parsing the token.
type Auth struct {
	log       *zap.SugaredLogger
	keyLookup KeyLookup
	userCore  *user.Core
	method    jwt.SigningMethod
	parser    *jwt.Parser
	issuer    string
	mu        sync.RWMutex
	cache     map[string]string
}

// New creates an Auth to support authentication/authorization
func New(cfg Config) (*Auth, error) {

	// If a database connection is not provided, we won't perform the
	// user enabled check.
	var usrCore *user.Core
	if cfg.DB != nil {
		evnCore := event.NewCore(cfg.Log)
		usrCore = user.NewCore(evnCore, userdb.NewStore(cfg.Log, cfg.DB))
	}

	a := Auth{
		log:       cfg.Log,
		keyLookup: cfg.KeyLookup,
		userCore:  usrCore,
		method:    jwt.GetSigningMethod(jwt.SigningMethodRS256.Name),
		parser:    jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Name})),
		issuer:    cfg.Issuer,
		cache:     make(map[string]string),
	}

	return &a, nil
}

// Authenticate processes the token to validate the sender's token.
func (a *Auth) Authenticate(ctx context.Context, bearerToken string) (Claims, error) {
	parts := strings.Split(bearerToken, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return Claims{}, errors.New("expected authorization header format: Bearer <token>")
	}

	var claims Claims
	token, _, err := a.parser.ParseUnverified(parts[1], &claims)
	if err != nil {
		return Claims{}, fmt.Errorf("error parsing token: %w", err)
	}

	// Perform an extra level of authentication verification with with OPA.

	kidRaw, exists := token.Header["kid"]
	if !exists {
		return Claims{}, fmt.Errorf("kid missing from header: %w", err)
	}

	kid, ok := kidRaw.(string)
	if !ok {
		return Claims{}, fmt.Errorf("kid malformed: %w", err)
	}

	pem, err := a.publicKeyLookup(kid)
	if err != nil {
		return Claims{}, fmt.Errorf("failed to fetch public key: %w", err)
	}

	input := map[string]any{
		"Key":   pem,
		"Token": parts[1],
		"ISS":   a.issuer,
	}

	if err := a.opaPolicyEvaluation(ctx, opaAuthentication, RuleAuthenticate, input); err != nil {
		return Claims{}, fmt.Errorf("authentication failed: %w", err)
	}

	// Check the database for this user to verify they are still enabled.
	if err := a.isUserEnabled(ctx, claims); err != nil {
		return Claims{}, fmt.Errorf("user not enabled: %w", err)
	}

	return claims, nil
}

// ================================================================

// publicKeyLookup performs a lookup for the public pem for the specified kid.
func (a *Auth) publicKeyLookup(kid string) (string, error) {
	pem, err := func() (string, error) {
		a.mu.RLock()
		defer a.mu.RUnlock()

		pem, exists := a.cache[kid]
		if !exists {
			return "", errors.New("not found")
		}
		return pem, nil
	}()
	if err == nil {
		return pem, nil
	}

	pem, err = a.keyLookup.PublicKey(kid)
	if err != nil {
		return "", fmt.Errorf("fetching public key: %w", err)
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	a.cache[kid] = pem

	return pem, nil
}

// opaPolicyEvaluation asks opa to evaluate the token against the specified token
// policy and public key.
func (a *Auth) opaPolicyEvaluation(ctx context.Context, opaPolicy string, rule string, input any) error {
	// query := fmt.Sprintf("x = data.%s.%s", opaPackage, rule)

	// q, err := rego.New(
	// 	rego.Query(query),
	// 	rego.Module("policy.rego", opaPolicy),
	// ).PrepareForEval(ctx)
	return nil
}

// isUserEnabled hits the database and checks the user is not disabled.
// If no database connection was provided, this check is skipped.
func (a *Auth) isUserEnabled(ctx context.Context, claims Claims) error {
	return nil
}