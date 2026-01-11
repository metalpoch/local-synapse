package authentication

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/valkey-io/valkey-go"
)

var (
	ErrEmptyUserID         = errors.New("empty user id")
	ErrEmptyRefreshToken   = errors.New("empty refresh token")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrInvalidAccessToken  = errors.New("invalid access token")
	ErrTokenGeneration     = errors.New("token generation failed")
	ErrTokenRefresh        = errors.New("token refresh failed")
	ErrTokenRevoked        = errors.New("token has been revoked")
	ErrTokenNotFound       = errors.New("token not found")
	ErrInvalidTokenFormat  = errors.New("invalid token format")
)

var luaReplaceSession = valkey.NewLuaScript(`
local userKey   = KEYS[1]
local newRTKey  = KEYS[2]
local newRT     = ARGV[1]
local userID    = ARGV[2]
local ttl       = ARGV[3]
local oldHash   = ARGV[4]

if oldHash and oldHash ~= "" then
    redis.call("DEL", "auth:rt:" .. oldHash)
end
redis.call("DEL", userKey)
redis.call("SET", newRTKey, userID, "EX", ttl)
redis.call("SET", userKey, newRT, "EX", ttl)
return 1
`)

var refreshCheck = valkey.NewLuaScript(`
-- refresh_check.lua
local oldRTKey = KEYS[1]
local userKey  = KEYS[2]
local oldRT    = ARGV[1]

local userID = redis.call("GET", oldRTKey)
if not userID then
    return -1  -- ErrInvalidRefreshToken
end

local current = redis.call("GET", userKey)
if current ~= oldRT then
    -- Detectamos intento de reúso: limpiamos la sesión por seguridad
    redis.call("DEL", userKey) 
    return -2  -- ErrTokenRevoked
end

return userID
`)

type AuthManager interface {
	GenerateTokens(ctx context.Context, userID string) (string, string, error)
	RefreshToken(ctx context.Context, oldRefreshToken string) (string, string, error)
	Logout(ctx context.Context, accessToken, refreshToken string) error
	ValidateAccessToken(ctx context.Context, tokenStr string) (string, error)
	GetSecret() []byte
}

type authManager struct {
	secret          []byte
	cache           valkey.Client
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewAuthManager(secret []byte, cache valkey.Client, accessTokenTTL, refreshTokenTTL time.Duration) *authManager {
	return &authManager{secret, cache, accessTokenTTL, refreshTokenTTL}
}

func (a *authManager) GetSecret() []byte { return a.secret }

func (a *authManager) GenerateTokens(ctx context.Context, userID string) (string, string, error) {
	if userID == "" {
		return "", "", ErrEmptyUserID
	}

	newRT, err := a.generateSecureToken(64)
	if err != nil {
		return "", "", fmt.Errorf("%w: %v", ErrTokenGeneration, err)
	}

	newRTKey := a.refreshTokenKey(newRT)
	userKey := a.userSessionKey(userID)
	ttlSec := int64(a.refreshTokenTTL.Seconds())

	var oldRTHash string
	oldRTPlain, err := a.cache.Do(ctx, a.cache.B().Get().Key(userKey).Build()).ToString()
	if err == nil && oldRTPlain != "" {
		h := sha256.Sum256([]byte(oldRTPlain))
		oldRTHash = hex.EncodeToString(h[:])
	}

	if err := luaReplaceSession.Exec(ctx, a.cache,
		[]string{userKey, newRTKey},
		[]string{newRT, userID, strconv.FormatInt(ttlSec, 10), oldRTHash},
	).Error(); err != nil {
		return "", "", fmt.Errorf("%w: lua execution failed: %v", ErrTokenGeneration, err)
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, &Claims{
		Sub: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(a.accessTokenTTL)),
		},
	}).SignedString(a.secret)

	if err != nil {
		for _, r := range a.cache.DoMulti(ctx,
			a.cache.B().Del().Key(newRTKey).Build(),
			a.cache.B().Del().Key(userKey).Build(),
		) {
			_ = r.Error()
		}
		return "", "", fmt.Errorf("%w: %v", ErrTokenGeneration, err)
	}

	return accessToken, newRT, nil
}

func (a *authManager) RefreshToken(ctx context.Context, oldRefreshToken string) (string, string, error) {
	if oldRefreshToken == "" {
		return "", "", ErrEmptyRefreshToken
	}

	oldKey := a.refreshTokenKey(oldRefreshToken)
	userID, _ := a.cache.Do(ctx, a.cache.B().Get().Key(oldKey).Build()).ToString()
	if userID == "" {
		return "", "", ErrInvalidRefreshToken
	}
	userKey := a.userSessionKey(userID)

	// el script devuelve el userID o nil/-1/-2
	ret, err := refreshCheck.Exec(ctx, a.cache,
		[]string{oldKey, userKey},
		[]string{oldRefreshToken},
	).ToString()

	switch {
	case err != nil:
		return "", "", fmt.Errorf("%w: %v", ErrTokenRefresh, err)
	case ret == "-1":
		return "", "", ErrInvalidRefreshToken
	case ret == "-2":
		return "", "", ErrTokenRevoked
	}

	// ret == userID
	return a.GenerateTokens(ctx, ret)
}

func (a *authManager) Logout(ctx context.Context, accessToken, refreshToken string) error {
	if accessToken == "" {
		return ErrInvalidAccessToken
	}

	token, _, err := jwt.NewParser().ParseUnverified(accessToken, &Claims{})
	if err != nil {
		return fmt.Errorf("failed to parse JWT: %w", err)
	}
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return ErrInvalidTokenFormat
	}

	exp, err := claims.GetExpirationTime()
	if err == nil {
		if ttl := time.Until(exp.Time); ttl > 0 {
			_ = a.cache.Do(ctx, a.cache.B().Set().
				Key(a.jwtBlacklistKey(accessToken)).
				Value("1").
				ExSeconds(int64(ttl.Seconds())).
				Build())
		}
	}

	userKey := a.userSessionKey(claims.Sub)

	if refreshToken == "" {
		refreshToken, _ = a.cache.Do(ctx, a.cache.B().Get().Key(userKey).Build()).ToString()
	}

	keysToDel := []string{userKey}
	if refreshToken != "" {
		keysToDel = append(keysToDel, a.refreshTokenKey(refreshToken))
	}

	return a.cache.Do(ctx, a.cache.B().Del().Key(keysToDel...).Build()).Error()
}

func (a *authManager) ValidateAccessToken(ctx context.Context, tokenStr string) (string, error) {
	if tokenStr == "" {
		return "", ErrInvalidAccessToken
	}

	blacklisted, err := a.isTokenBlacklisted(ctx, tokenStr)
	if err != nil {
		return "", err
	}
	if blacklisted {
		return "", ErrTokenRevoked
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return a.secret, nil
	})

	if err != nil || !token.Valid {
		return "", ErrInvalidAccessToken
	}

	return claims.Sub, nil
}

func (a *authManager) generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate secure token: %w", err)
	}

	return hex.EncodeToString(bytes), nil
}

func (a *authManager) userSessionKey(userID string) string {
	return "auth:user_session:" + userID
}

func (a *authManager) refreshTokenKey(rt string) string {
	hash := sha256.Sum256([]byte(rt))
	hashedToken := hex.EncodeToString(hash[:])
	return "auth:rt:" + hashedToken
}

func (a *authManager) jwtBlacklistKey(jwtToken string) string {
	hash := sha256.Sum256([]byte(jwtToken))
	hashedToken := hex.EncodeToString(hash[:])
	return "auth:jwt:blacklist:" + hashedToken
}

func (a *authManager) isTokenBlacklisted(ctx context.Context, jwtToken string) (bool, error) {
	if jwtToken == "" {
		return false, ErrInvalidAccessToken
	}

	jwtKey := a.jwtBlacklistKey(jwtToken)
	exists, err := a.cache.Do(ctx, a.cache.B().Exists().Key(jwtKey).Build()).AsBool()
	if err != nil {
		return false, fmt.Errorf("failed to check token blacklist: %w", err)
	}

	return exists, nil
}
