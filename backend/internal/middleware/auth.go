package middleware

import (
	"context"
	"crypto/subtle"
	"fmt"
	"strings"

	"connectrpc.com/connect"
)

// TokenValidator validates a Bearer token. Implement this interface to add
// JWT, OIDC, or API-key validation in the future.
type TokenValidator interface {
	Validate(token string) error
}

// StaticTokenValidator accepts a single pre-configured token.
// Suitable for simple self-hosted single-user deployments.
type StaticTokenValidator struct{ token string }

func NewStaticTokenValidator(token string) TokenValidator {
	return &StaticTokenValidator{token: token}
}

func (v *StaticTokenValidator) Validate(token string) error {
	// Constant-time comparison to prevent timing attacks.
	if subtle.ConstantTimeCompare([]byte(token), []byte(v.token)) != 1 {
		return connect.NewError(connect.CodeUnauthenticated, nil)
	}
	return nil
}

// AuthInterceptor is a connect-go interceptor that optionally enforces auth.
// When enabled=false it is a no-op, allowing the server to run without auth.
type AuthInterceptor struct {
	enabled   bool
	validator TokenValidator
}

// NewAuthInterceptor creates the interceptor. Pass enabled=false (the default)
// to bypass auth entirely. Pass a TokenValidator to enforce token checking.
func NewAuthInterceptor(enabled bool, validator TokenValidator) *AuthInterceptor {
	return &AuthInterceptor{enabled: enabled, validator: validator}
}

// WrapUnary implements connect.Interceptor.
func (a *AuthInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		if !a.enabled {
			return next(ctx, req)
		}
		if err := a.checkAuth(req.Header().Get("Authorization")); err != nil {
			return nil, err
		}
		return next(ctx, req)
	}
}

// WrapStreamingClient implements connect.Interceptor (pass-through).
func (a *AuthInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

// WrapStreamingHandler implements connect.Interceptor.
func (a *AuthInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		if !a.enabled {
			return next(ctx, conn)
		}
		if err := a.checkAuth(conn.RequestHeader().Get("Authorization")); err != nil {
			return err
		}
		return next(ctx, conn)
	}
}

func (a *AuthInterceptor) checkAuth(authHeader string) error {
	if a.validator == nil {
		// Auth enabled but no validator configured — fail closed rather than admit all.
		return connect.NewError(connect.CodeInternal, fmt.Errorf("auth enabled but no validator configured"))
	}
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return connect.NewError(connect.CodeUnauthenticated, nil)
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")
	return a.validator.Validate(token)
}
