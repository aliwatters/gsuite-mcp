package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/aliwatters/gsuite-mcp/internal/auth"
	"github.com/aliwatters/gsuite-mcp/internal/config"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	"google.golang.org/api/people/v1"
	"google.golang.org/api/sheets/v4"
	"google.golang.org/api/tasks/v1"
)

// apiCheck defines a single API check to perform.
type apiCheck struct {
	name      string
	apiID     string
	checkFunc func(ctx context.Context, client *http.Client) error
}

// runCheck performs preflight validation of configuration, accounts, and API access.
func runCheck() {
	issues := 0

	// Stage 1: Configuration
	fmt.Println("Checking configuration...")

	mgr, projectNumber, err := checkConfiguration()
	if err != nil {
		fmt.Printf("  ✗ %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("  ✓ client_secret.json found\n")
	if projectNumber != "" {
		fmt.Printf("  ✓ OAuth client ID (project: %s)\n", projectNumber)
	}

	port, envOverride, err := auth.ResolveOAuthPort()
	if err != nil {
		fmt.Printf("  ✗ OAuth port: %v\n", err)
		issues++
	} else if envOverride {
		fmt.Printf("  ✓ OAuth port: %d (env override)\n", port)
	} else {
		fmt.Printf("  ✓ OAuth port: %d\n", port)
	}

	// Stage 2: Accounts
	fmt.Println("\nChecking accounts...")

	emails, err := config.GetAuthenticatedEmails()
	if err != nil {
		fmt.Printf("  ✗ Error reading credentials: %v\n", err)
		os.Exit(1)
	}
	if len(emails) == 0 {
		fmt.Printf("  ✗ No authenticated accounts found\n")
		fmt.Printf("    Run 'gsuite-mcp auth' to authenticate a Google account.\n")
		os.Exit(1)
	}

	ctx := context.Background()
	var validEmails []string

	for _, email := range emails {
		_, err := mgr.GetClientForEmail(ctx, email)
		if err != nil {
			fmt.Printf("  ✗ %s — token refresh failed: %v\n", email, err)
			fmt.Printf("    Run 'gsuite-mcp auth' and sign in with this account to fix.\n")
			issues++
			continue
		}
		fmt.Printf("  ✓ %s — token valid\n", email)
		validEmails = append(validEmails, email)
	}

	// Stage 3: API access
	checks := buildAPIChecks()

	for _, email := range validEmails {
		fmt.Printf("\nChecking API access for %s...\n", email)

		client, err := mgr.GetClientForEmail(ctx, email)
		if err != nil {
			fmt.Printf("  ✗ %s — failed to get client: %v\n", email, err)
			issues++
			continue
		}

		for _, check := range checks {
			err := check.checkFunc(ctx, client)
			if err == nil {
				fmt.Printf("  ✓ %s\n", check.name)
				continue
			}

			if isAPIDisabled(err) {
				fmt.Printf("  ✗ %s — not enabled\n", check.name)
				if projectNumber != "" {
					fmt.Printf("    Enable at: https://console.developers.google.com/apis/api/%s/overview?project=%s\n", check.apiID, projectNumber)
				}
				issues++
			} else {
				fmt.Printf("  ✗ %s — %v\n", check.name, err)
				issues++
			}
		}
	}

	// Summary
	fmt.Println()
	if issues == 0 {
		fmt.Println("All checks passed!")
	} else {
		fmt.Printf("%d issue(s) found.\n", issues)
		os.Exit(1)
	}
}

// checkConfiguration validates client_secret.json and extracts the GCP project number.
func checkConfiguration() (*auth.Manager, string, error) {
	mgr, err := auth.NewManager()
	if err != nil {
		return nil, "", err
	}

	projectNumber := extractProjectNumber(mgr.OAuthConfig().ClientID)
	return mgr, projectNumber, nil
}

// extractProjectNumber extracts the GCP project number from a client ID.
// Client IDs have the format: <project-number>-<random>.apps.googleusercontent.com
func extractProjectNumber(clientID string) string {
	if idx := strings.Index(clientID, "-"); idx > 0 {
		candidate := clientID[:idx]
		// Verify it looks like a number
		for _, c := range candidate {
			if c < '0' || c > '9' {
				return ""
			}
		}
		return candidate
	}
	return ""
}

// isAPIDisabled checks if an error indicates that a Google API is not enabled.
func isAPIDisabled(err error) bool {
	var apiErr *googleapi.Error
	if !errors.As(err, &apiErr) {
		return false
	}
	if apiErr.Code != 403 {
		return false
	}
	msg := apiErr.Message
	return strings.Contains(msg, "has not been used") ||
		strings.Contains(msg, "is disabled") ||
		strings.Contains(msg, "it is disabled")
}

// apiCheckFunc is a function that performs an API health check.
type apiCheckFunc[S any] func(ctx context.Context, client *http.Client) (S, error)

// makeAPICheck creates an apiCheck using a service constructor and a probe function.
// The probe returns an error if the API is unavailable or the call fails.
// If allowNotFound is true, a 404 response is treated as success (API is enabled,
// the resource simply doesn't exist).
func makeAPICheck[S any](name, apiID string, newSrv apiCheckFunc[S], probe func(ctx context.Context, srv S) error, allowNotFound bool) apiCheck {
	return apiCheck{
		name:  name,
		apiID: apiID,
		checkFunc: func(ctx context.Context, client *http.Client) error {
			srv, err := newSrv(ctx, client)
			if err != nil {
				return fmt.Errorf("creating %s service: %w", name, err)
			}
			err = probe(ctx, srv)
			if allowNotFound && isNotFound(err) {
				return nil
			}
			if err != nil {
				return fmt.Errorf("%s probe: %w", name, err)
			}
			return nil
		},
	}
}

// buildAPIChecks returns the list of API checks to perform.
func buildAPIChecks() []apiCheck {
	return []apiCheck{
		makeAPICheck("Gmail API", "gmail.googleapis.com",
			func(ctx context.Context, client *http.Client) (*gmail.Service, error) {
				return gmail.NewService(ctx, option.WithHTTPClient(client))
			},
			func(ctx context.Context, srv *gmail.Service) error {
				_, err := srv.Users.GetProfile("me").Fields("emailAddress").Do()
				return err
			}, false),

		makeAPICheck("Google Calendar API", "calendar-json.googleapis.com",
			func(ctx context.Context, client *http.Client) (*calendar.Service, error) {
				return calendar.NewService(ctx, option.WithHTTPClient(client))
			},
			func(ctx context.Context, srv *calendar.Service) error {
				_, err := srv.CalendarList.List().MaxResults(1).Fields("items(id)").Do()
				return err
			}, false),

		makeAPICheck("Google Drive API", "drive.googleapis.com",
			func(ctx context.Context, client *http.Client) (*drive.Service, error) {
				return drive.NewService(ctx, option.WithHTTPClient(client))
			},
			func(ctx context.Context, srv *drive.Service) error {
				_, err := srv.Files.List().PageSize(1).Fields("files(id)").Do()
				return err
			}, false),

		// 404 means the API is enabled (document doesn't exist, which is expected)
		makeAPICheck("Google Docs API", "docs.googleapis.com",
			func(ctx context.Context, client *http.Client) (*docs.Service, error) {
				return docs.NewService(ctx, option.WithHTTPClient(client))
			},
			func(ctx context.Context, srv *docs.Service) error {
				_, err := srv.Documents.Get("_check").Fields("title").Do()
				return err
			}, true),

		// 404 means the API is enabled (spreadsheet doesn't exist, which is expected)
		makeAPICheck("Google Sheets API", "sheets.googleapis.com",
			func(ctx context.Context, client *http.Client) (*sheets.Service, error) {
				return sheets.NewService(ctx, option.WithHTTPClient(client))
			},
			func(ctx context.Context, srv *sheets.Service) error {
				_, err := srv.Spreadsheets.Get("_check").Fields("spreadsheetId").Do()
				return err
			}, true),

		makeAPICheck("Google Tasks API", "tasks.googleapis.com",
			func(ctx context.Context, client *http.Client) (*tasks.Service, error) {
				return tasks.NewService(ctx, option.WithHTTPClient(client))
			},
			func(ctx context.Context, srv *tasks.Service) error {
				_, err := srv.Tasklists.List().MaxResults(1).Do()
				return err
			}, false),

		makeAPICheck("People API (Contacts)", "people.googleapis.com",
			func(ctx context.Context, client *http.Client) (*people.Service, error) {
				return people.NewService(ctx, option.WithHTTPClient(client))
			},
			func(ctx context.Context, srv *people.Service) error {
				_, err := srv.People.Connections.List("people/me").PageSize(1).PersonFields("names").Do()
				return err
			}, false),
	}
}

// isNotFound checks if an error is a 404 Not Found response.
func isNotFound(err error) bool {
	var apiErr *googleapi.Error
	return errors.As(err, &apiErr) && apiErr.Code == 404
}
