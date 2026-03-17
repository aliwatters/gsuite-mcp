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

// buildAPIChecks returns the list of API checks to perform.
func buildAPIChecks() []apiCheck {
	return []apiCheck{
		{
			name:  "Gmail API",
			apiID: "gmail.googleapis.com",
			checkFunc: func(ctx context.Context, client *http.Client) error {
				srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
				if err != nil {
					return err
				}
				_, err = srv.Users.GetProfile("me").Fields("emailAddress").Do()
				return err
			},
		},
		{
			name:  "Google Calendar API",
			apiID: "calendar-json.googleapis.com",
			checkFunc: func(ctx context.Context, client *http.Client) error {
				srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
				if err != nil {
					return err
				}
				_, err = srv.CalendarList.List().MaxResults(1).Fields("items(id)").Do()
				return err
			},
		},
		{
			name:  "Google Drive API",
			apiID: "drive.googleapis.com",
			checkFunc: func(ctx context.Context, client *http.Client) error {
				srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
				if err != nil {
					return err
				}
				_, err = srv.Files.List().PageSize(1).Fields("files(id)").Do()
				return err
			},
		},
		{
			name:  "Google Docs API",
			apiID: "docs.googleapis.com",
			checkFunc: func(ctx context.Context, client *http.Client) error {
				srv, err := docs.NewService(ctx, option.WithHTTPClient(client))
				if err != nil {
					return err
				}
				_, err = srv.Documents.Get("_check").Fields("title").Do()
				// 404 means the API is enabled (document doesn't exist, which is expected)
				if isNotFound(err) {
					return nil
				}
				return err
			},
		},
		{
			name:  "Google Sheets API",
			apiID: "sheets.googleapis.com",
			checkFunc: func(ctx context.Context, client *http.Client) error {
				srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
				if err != nil {
					return err
				}
				_, err = srv.Spreadsheets.Get("_check").Fields("spreadsheetId").Do()
				// 404 means the API is enabled (spreadsheet doesn't exist, which is expected)
				if isNotFound(err) {
					return nil
				}
				return err
			},
		},
		{
			name:  "Google Tasks API",
			apiID: "tasks.googleapis.com",
			checkFunc: func(ctx context.Context, client *http.Client) error {
				srv, err := tasks.NewService(ctx, option.WithHTTPClient(client))
				if err != nil {
					return err
				}
				_, err = srv.Tasklists.List().MaxResults(1).Do()
				return err
			},
		},
		{
			name:  "People API (Contacts)",
			apiID: "people.googleapis.com",
			checkFunc: func(ctx context.Context, client *http.Client) error {
				srv, err := people.NewService(ctx, option.WithHTTPClient(client))
				if err != nil {
					return err
				}
				_, err = srv.People.Connections.List("people/me").PageSize(1).PersonFields("names").Do()
				return err
			},
		},
	}
}

// isNotFound checks if an error is a 404 Not Found response.
func isNotFound(err error) bool {
	var apiErr *googleapi.Error
	return errors.As(err, &apiErr) && apiErr.Code == 404
}
