package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/plexreader/plexreader/backend/internal/middleware"
)

// exerciseInterceptor builds a minimal connect server with the interceptor and
// sends a request with the given Authorization header value (empty = omit).
// Returns the HTTP status code and any connect error code.
func exerciseInterceptor(t *testing.T, interceptor *middleware.AuthInterceptor, authHeader string) int {
	t.Helper()
	mux := http.NewServeMux()

	// Use a trivial handler that just returns OK.
	type noopHandler struct{}
	// We can't use a real proto service without generated code here, so we
	// build a raw path handler that simulates a connect unary handler result.
	// Instead, test the interceptor's WrapUnary directly via a real call.
	_ = mux
	_ = noopHandler{}

	called := false
	next := func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		called = true
		return nil, nil
	}

	// Build a fake request using httptest to get real headers into a real
	// http.Request, then extract the Authorization for checkAuth via
	// a lightweight custom AnyRequest wrapper.
	//
	// Since connect.AnyRequest is sealed, we call WrapUnary and pass a
	// real connect.Request (but we need to use the handler approach).
	// Simplest: call the WrapUnary wrapping directly with a concrete
	// *connect.Request[emptypb.Empty] which is a real AnyRequest.
	inner := connect.NewRequest(&emptypb.Empty{})
	if authHeader != "" {
		inner.Header().Set("Authorization", authHeader)
	}
	_, err := interceptor.WrapUnary(next)(context.Background(), inner)

	// Translate to status code.
	if err == nil {
		if !called {
			t.Error("handler not called despite no error")
		}
		return http.StatusOK
	}
	var ce *connect.Error
	if connect.IsNotModifiedError(err) {
		return http.StatusNotModified
	}
	if ok := false; !ok {
		if ce, ok = err.(*connect.Error); !ok {
			t.Fatalf("expected connect.Error, got %T: %v", err, err)
		}
	}
	switch ce.Code() {
	case connect.CodeUnauthenticated:
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}

// helper to create a test server — kept simple by exercising WrapUnary directly.
func makeInterceptor(enabled bool, token string) *middleware.AuthInterceptor {
	var v middleware.TokenValidator
	if token != "" {
		v = middleware.NewStaticTokenValidator(token)
	}
	return middleware.NewAuthInterceptor(enabled, v)
}

func TestAuthDisabled_PassThrough(t *testing.T) {
	i := makeInterceptor(false, "")
	called := false
	next := func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		called = true
		return nil, nil
	}
	req := connect.NewRequest(&emptypb.Empty{})
	// no Authorization header
	_, err := i.WrapUnary(next)(context.Background(), req)
	if err != nil || !called {
		t.Errorf("disabled interceptor should pass through: err=%v called=%v", err, called)
	}
}

func TestAuthEnabled_MissingHeader(t *testing.T) {
	i := makeInterceptor(true, "secret")
	req := connect.NewRequest(&emptypb.Empty{})
	_, err := i.WrapUnary(passHandler(t))(context.Background(), req)
	assertUnauthenticated(t, err)
}

func TestAuthEnabled_WrongToken(t *testing.T) {
	i := makeInterceptor(true, "secret")
	req := connect.NewRequest(&emptypb.Empty{})
	req.Header().Set("Authorization", "Bearer wrongtoken")
	_, err := i.WrapUnary(passHandler(t))(context.Background(), req)
	assertUnauthenticated(t, err)
}

func TestAuthEnabled_CorrectToken(t *testing.T) {
	i := makeInterceptor(true, "secret")
	called := false
	next := func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		called = true
		return nil, nil
	}
	req := connect.NewRequest(&emptypb.Empty{})
	req.Header().Set("Authorization", "Bearer secret")
	_, err := i.WrapUnary(next)(context.Background(), req)
	if err != nil || !called {
		t.Errorf("valid token rejected: err=%v called=%v", err, called)
	}
}

func TestAuthEnabled_NoBearerPrefix(t *testing.T) {
	i := makeInterceptor(true, "secret")
	req := connect.NewRequest(&emptypb.Empty{})
	req.Header().Set("Authorization", "secret") // missing "Bearer " prefix
	_, err := i.WrapUnary(passHandler(t))(context.Background(), req)
	assertUnauthenticated(t, err)
}

func TestAuthEnabled_BasicPrefixRejected(t *testing.T) {
	i := makeInterceptor(true, "secret")
	req := connect.NewRequest(&emptypb.Empty{})
	req.Header().Set("Authorization", "Basic dXNlcjpwYXNz")
	_, err := i.WrapUnary(passHandler(t))(context.Background(), req)
	assertUnauthenticated(t, err)
}

func TestStaticValidator_ConstantTimeComparison(t *testing.T) {
	// Smoke test: ensure different-length tokens don't panic.
	v := middleware.NewStaticTokenValidator("short")
	if err := v.Validate("much-longer-token-than-secret"); err == nil {
		t.Error("longer wrong token should fail")
	}
	if err := v.Validate("short"); err != nil {
		t.Errorf("correct token rejected: %v", err)
	}
}

func TestAuthEnabled_NilValidator_FailsClosed(t *testing.T) {
	// When auth is enabled but no validator is configured, requests must be
	// rejected with CodeInternal rather than let through (fail-closed).
	i := middleware.NewAuthInterceptor(true, nil)
	req := connect.NewRequest(&emptypb.Empty{})
	req.Header().Set("Authorization", "Bearer anything")
	_, err := i.WrapUnary(passHandler(t))(context.Background(), req)
	if err == nil {
		t.Fatal("expected error with nil validator, got nil")
	}
	ce, ok := err.(*connect.Error)
	if !ok {
		t.Fatalf("expected *connect.Error, got %T: %v", err, err)
	}
	if ce.Code() != connect.CodeInternal {
		t.Errorf("expected CodeInternal for nil validator, got %v", ce.Code())
	}
}

// passHandler returns a next func that fails the test if called.
func passHandler(t *testing.T) connect.UnaryFunc {
	t.Helper()
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		t.Helper()
		t.Error("handler must not be called when auth fails")
		return nil, nil
	}
}

func assertUnauthenticated(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected unauthenticated error, got nil")
	}
	ce, ok := err.(*connect.Error)
	if !ok {
		t.Fatalf("expected *connect.Error, got %T: %v", err, err)
	}
	if ce.Code() != connect.CodeUnauthenticated {
		t.Errorf("expected CodeUnauthenticated, got %v", ce.Code())
	}
}

// Ensure the httptest import is used — used for future integration tests.
var _ = httptest.NewServer
var _ = strings.Contains
