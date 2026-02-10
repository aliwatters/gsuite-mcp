package common

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mark3labs/mcp-go/mcp"
)

// ServiceFactory creates services of type S for a given email.
type ServiceFactory[S any] interface {
	CreateService(ctx context.Context, email string) (S, error)
}

// HandlerDeps contains dependencies for tool handlers, parameterized by service type.
// This allows for dependency injection in tests.
type HandlerDeps[S any] struct {
	EmailResolver  func(request mcp.CallToolRequest) (string, error)
	ServiceFactory ServiceFactory[S]
}

// getServiceWithDeps resolves the email from the request and creates a service.
func getServiceWithDeps[S any](ctx context.Context, request mcp.CallToolRequest, deps *HandlerDeps[S], defaultDeps *HandlerDeps[S]) (S, error) {
	if deps == nil {
		deps = defaultDeps
	}

	email, err := deps.EmailResolver(request)
	if err != nil {
		var zero S
		return zero, err
	}

	return deps.ServiceFactory.CreateService(ctx, email)
}

// MarshalToolResult marshals the result to JSON and returns an MCP tool result.
func MarshalToolResult(result any) (*mcp.CallToolResult, error) {
	jsonData, err := json.Marshal(result)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error encoding response: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

// ServiceConstructor creates a service S from a context and HTTP client.
type ServiceConstructor[S any] func(ctx context.Context, client *http.Client) (S, error)

// NewDefaultHandlerDeps creates production HandlerDeps using ResolveAccountFromRequest
// and a lazyServiceFactory backed by the global auth manager.
func NewDefaultHandlerDeps[S any](constructor ServiceConstructor[S]) *HandlerDeps[S] {
	return &HandlerDeps[S]{
		EmailResolver: ResolveAccountFromRequest,
		ServiceFactory: &lazyServiceFactory[S]{
			Constructor: constructor,
		},
	}
}

// lazyServiceFactory resolves the auth manager lazily at call time (not init time)
// since common.GetDeps() is nil during package initialization.
type lazyServiceFactory[S any] struct {
	Constructor ServiceConstructor[S]
}

func (f *lazyServiceFactory[S]) CreateService(ctx context.Context, email string) (S, error) {
	d := GetDeps()
	if d == nil || d.AuthManager == nil {
		var zero S
		return zero, fmt.Errorf("authentication subsystem not initialized; restart gsuite-mcp and check server startup logs")
	}
	client, err := d.AuthManager.GetClientOrAuthenticate(ctx, email, false)
	if err != nil {
		var zero S
		return zero, err
	}
	return f.Constructor(ctx, client)
}

// ResolveServiceOrError resolves a service from the request and deps, returning
// an MCP error result if resolution fails. Returns (service, nil, true) on success
// or (zero, errorResult, false) on failure.
func ResolveServiceOrError[S any](ctx context.Context, request mcp.CallToolRequest, deps *HandlerDeps[S], defaultDeps *HandlerDeps[S]) (S, *mcp.CallToolResult, bool) {
	svc, err := getServiceWithDeps(ctx, request, deps, defaultDeps)
	if err != nil {
		var zero S
		return zero, mcp.NewToolResultError(err.Error()), false
	}
	return svc, nil, true
}

// MockServiceFactory creates mock services for testing.
type MockServiceFactory[S any] struct {
	MockService S
}

// CreateService returns the mock service.
func (f *MockServiceFactory[S]) CreateService(ctx context.Context, email string) (S, error) {
	return f.MockService, nil
}

// testableFunc is a handler function that accepts deps for dependency injection.
type testableFunc[S any] func(ctx context.Context, request mcp.CallToolRequest, deps *HandlerDeps[S]) (*mcp.CallToolResult, error)

// WrapHandler wraps a testableFunc into a standard MCP handler by passing nil deps,
// which causes the testable function to resolve production dependencies.
func WrapHandler[S any](fn testableFunc[S]) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return fn(ctx, request, nil)
	}
}

// TestEmail is the default email used across all test fixtures.
const TestEmail = "test@example.com"

// TestFixtures contains the common test fixture components for any service type.
type TestFixtures[S any] struct {
	DefaultEmail string
	MockService  S
	Deps         *HandlerDeps[S]
}

// NewTestFixtures creates test fixtures with a standard email resolver and mock service factory.
func NewTestFixtures[S any](mockService S) *TestFixtures[S] {
	defaultEmail := TestEmail
	emailResolver := func(request mcp.CallToolRequest) (string, error) {
		if accountParam, ok := request.Params.Arguments["account"].(string); ok && accountParam != "" {
			return accountParam, nil
		}
		return defaultEmail, nil
	}

	deps := &HandlerDeps[S]{
		EmailResolver: emailResolver,
		ServiceFactory: &MockServiceFactory[S]{
			MockService: mockService,
		},
	}

	return &TestFixtures[S]{
		DefaultEmail: defaultEmail,
		MockService:  mockService,
		Deps:         deps,
	}
}
