package main

import (
	"context"
	"encoding/json"
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

// checkExitCode defines exit code semantics for the check command.
// These are stable for use in cron jobs and automation scripts.
const (
	checkExitOK          = 0 // all checks passed
	checkExitStale       = 1 // one or more tokens are stale / need re-auth
	checkExitConfigError = 2 // configuration error (missing client_secret.json, etc.)
)

// apiCheck defines a single API check to perform.
type apiCheck struct {
	name      string
	apiID     string
	checkFunc func(ctx context.Context, client *http.Client) error
}

// checkAccountResult holds the per-account result of a check run.
type checkAccountResult struct {
	Email      string            `json:"email"`
	TokenValid bool              `json:"token_valid"`
	TokenError string            `json:"token_error,omitempty"`
	APIs       []checkAPIResult  `json:"apis,omitempty"`
}

// checkAPIResult holds the per-API result for a single account.
type checkAPIResult struct {
	Name    string `json:"name"`
	OK      bool   `json:"ok"`
	Error   string `json:"error,omitempty"`
	HelpURL string `json:"help_url,omitempty"`
}

// checkReport is the full structured output of a check run.
type checkReport struct {
	OK            bool                 `json:"ok"`
	ConfigOK      bool                 `json:"config_ok"`
	ConfigError   string               `json:"config_error,omitempty"`
	OAuthPort     int                  `json:"oauth_port,omitempty"`
	ProjectNumber string               `json:"project_number,omitempty"`
	Accounts      []checkAccountResult `json:"accounts"`
	Issues        int                  `json:"issues"`
}

// runCheck performs preflight validation of configuration, accounts, and API access.
// It checks for --json in os.Args to select output mode.
func runCheck() {
	jsonMode := false
	for _, arg := range os.Args[2:] {
		if arg == "--json" {
			jsonMode = true
		}
	}

	report := checkReport{
		Accounts: []checkAccountResult{},
	}

	// Stage 1: Configuration
	mgr, projectNumber, configErr := checkConfiguration()
	if configErr != nil {
		report.ConfigOK = false
		report.ConfigError = configErr.Error()
		if jsonMode {
			outputJSON(report)
		} else {
			fmt.Fprintf(os.Stderr, "Configuration error: %v\n", configErr)
		}
		os.Exit(checkExitConfigError)
	}
	report.ConfigOK = true
	report.ProjectNumber = projectNumber

	port, _, portErr := auth.ResolveOAuthPort()
	if portErr != nil {
		report.ConfigOK = false
		report.ConfigError = fmt.Sprintf("OAuth port: %v", portErr)
		if jsonMode {
			outputJSON(report)
		} else {
			fmt.Println("Checking configuration...")
			fmt.Printf("  ✓ client_secret.json found\n")
			if projectNumber != "" {
				fmt.Printf("  ✓ OAuth client ID (project: %s)\n", projectNumber)
			}
			fmt.Printf("  ✗ OAuth port: %v\n", portErr)
		}
		os.Exit(checkExitConfigError)
	}
	report.OAuthPort = port

	if !jsonMode {
		fmt.Println("Checking configuration...")
		fmt.Printf("  ✓ client_secret.json found\n")
		if projectNumber != "" {
			fmt.Printf("  ✓ OAuth client ID (project: %s)\n", projectNumber)
		}
		fmt.Printf("  ✓ OAuth port: %d\n", port)
	}

	// Stage 2: Accounts
	emails, err := config.GetAuthenticatedEmails()
	if err != nil {
		if jsonMode {
			report.ConfigError = fmt.Sprintf("reading credentials: %v", err)
			outputJSON(report)
		} else {
			fmt.Fprintf(os.Stderr, "Error reading credentials: %v\n", err)
		}
		os.Exit(checkExitConfigError)
	}
	if len(emails) == 0 {
		if jsonMode {
			report.ConfigError = "no authenticated accounts found"
			outputJSON(report)
		} else {
			fmt.Fprintf(os.Stderr, "No authenticated accounts found.\n")
			fmt.Fprintf(os.Stderr, "Run 'gsuite-mcp auth' to authenticate a Google account.\n")
		}
		os.Exit(checkExitStale)
	}

	if !jsonMode {
		fmt.Println("\nChecking accounts...")
	}

	ctx := context.Background()
	var validEmails []string
	staleCount := 0
	// accountIndex maps email → index in report.Accounts for API result attachment.
	accountIndex := make(map[string]int, len(emails))

	for _, email := range emails {
		acct := checkAccountResult{Email: email}
		_, tokenErr := mgr.GetClientForEmail(ctx, email)
		if tokenErr != nil {
			acct.TokenValid = false
			acct.TokenError = tokenErr.Error()
			staleCount++
			if !jsonMode {
				fmt.Printf("  ✗ %s — token refresh failed: %v\n", email, tokenErr)
				fmt.Printf("    Run 'gsuite-mcp auth' and sign in with this account to fix.\n")
			}
		} else {
			acct.TokenValid = true
			validEmails = append(validEmails, email)
			if !jsonMode {
				fmt.Printf("  ✓ %s — token valid\n", email)
			}
		}
		accountIndex[email] = len(report.Accounts)
		report.Accounts = append(report.Accounts, acct)
	}

	// Stage 3: API access
	apiChecks := buildAPIChecks()
	apiIssues := 0

	for _, email := range validEmails {
		if !jsonMode {
			fmt.Printf("\nChecking API access for %s...\n", email)
		}

		idx := accountIndex[email]

		client, clientErr := mgr.GetClientForEmail(ctx, email)
		if clientErr != nil {
			if !jsonMode {
				fmt.Printf("  ✗ %s — failed to get client: %v\n", email, clientErr)
			}
			// Mark all APIs as failed for this account
			for _, check := range apiChecks {
				report.Accounts[idx].APIs = append(report.Accounts[idx].APIs, checkAPIResult{
					Name:  check.name,
					OK:    false,
					Error: clientErr.Error(),
				})
			}
			apiIssues++
			continue
		}

		for _, check := range apiChecks {
			apiErr := check.checkFunc(ctx, client)
			res := checkAPIResult{Name: check.name, OK: apiErr == nil}

			if apiErr != nil {
				apiIssues++
				if isAPIDisabled(apiErr) {
					res.Error = "API not enabled"
					if projectNumber != "" {
						res.HelpURL = fmt.Sprintf(
							"https://console.developers.google.com/apis/api/%s/overview?project=%s",
							check.apiID, projectNumber)
					}
					if !jsonMode {
						fmt.Printf("  ✗ %s — not enabled\n", check.name)
						if res.HelpURL != "" {
							fmt.Printf("    Enable at: %s\n", res.HelpURL)
						}
					}
				} else {
					res.Error = apiErr.Error()
					if !jsonMode {
						fmt.Printf("  ✗ %s — %v\n", check.name, apiErr)
					}
				}
			} else if !jsonMode {
				fmt.Printf("  ✓ %s\n", check.name)
			}

			report.Accounts[idx].APIs = append(report.Accounts[idx].APIs, res)
		}
	}

	totalIssues := staleCount + apiIssues
	report.Issues = totalIssues
	report.OK = totalIssues == 0

	if jsonMode {
		outputJSON(report)
	} else {
		fmt.Println()
		if totalIssues == 0 {
			fmt.Println("All checks passed!")
		} else {
			fmt.Printf("%d issue(s) found.\n", totalIssues)
		}
	}

	if staleCount > 0 {
		os.Exit(checkExitStale)
	}
	if totalIssues > 0 {
		os.Exit(checkExitStale)
	}
}

// outputJSON writes the check report as JSON to stdout.
func outputJSON(report checkReport) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(report); err != nil {
		fmt.Fprintf(os.Stderr, "error encoding JSON: %v\n", err)
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
