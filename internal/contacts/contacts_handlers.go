package contacts

import (
	"context"
	"net/http"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/option"
	"google.golang.org/api/people/v1"
)

// Type aliases using generic types from common package.
type ContactsHandlerDeps = common.HandlerDeps[ContactsService]
type ContactsServiceFactory = common.ServiceFactory[ContactsService]
type MockContactsServiceFactory = common.MockServiceFactory[ContactsService]

// NewContactsService creates a ContactsService from an authenticated HTTP client.
func NewContactsService(ctx context.Context, client *http.Client) (ContactsService, error) {
	srv, err := people.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}
	return NewRealContactsService(srv), nil
}

// DefaultContactsHandlerDeps holds the default dependencies for production use.
var DefaultContactsHandlerDeps = common.NewDefaultHandlerDeps(NewContactsService)

// ResolveContactsServiceOrError resolves a Contacts service, returning an MCP error result on failure.
func ResolveContactsServiceOrError(ctx context.Context, request mcp.CallToolRequest, deps *ContactsHandlerDeps) (ContactsService, *mcp.CallToolResult, bool) {
	return common.ResolveServiceOrError(ctx, request, deps, DefaultContactsHandlerDeps)
}
