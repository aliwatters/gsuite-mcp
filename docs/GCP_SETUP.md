# GCP Project Setup Guide

Step-by-step guide to creating and configuring a Google Cloud Platform project for gsuite-mcp. This guide reflects the current GCP Console UI as of 2026.

> **Already have a GCP project?** Skip to [Step 3: Enable APIs](#step-3-enable-the-required-apis) to make sure all required APIs are enabled.

---

## Overview

gsuite-mcp needs three things from GCP:

1. A **GCP project** to host your OAuth app
2. An **OAuth consent screen** so Google knows what your app is
3. **OAuth credentials** (Desktop app type) for authentication
4. **APIs enabled** for each Google service you want to use

Total time: ~10 minutes.

---

## Step 1: Create a GCP Project

**URL**: [https://console.cloud.google.com/projectcreate](https://console.cloud.google.com/projectcreate)

1. Go to the URL above (or click the project dropdown at the top of any GCP Console page, then "New Project")
2. Set the project name to something memorable, e.g., `gsuite-mcp`
3. Leave "Organization" and "Location" as defaults (for personal Gmail accounts, these are not shown)
4. Click **Create**
5. Wait for the notification that the project was created, then select it from the project dropdown

<!-- screenshot: create-project-dialog -->

> **Pitfall**: If you have multiple GCP projects, double-check the project selector at the top of the page shows your new project before proceeding. All subsequent steps must be done in the same project.

---

## Step 2: Configure the OAuth Consent Screen

### 2a: Accept Terms of Service (if prompted)

**URL**: [https://console.cloud.google.com/apis/credentials/consent](https://console.cloud.google.com/apis/credentials/consent)

If this is your first time using the GCP Console, you may see a Terms of Service prompt. Accept it before continuing.

<!-- screenshot: tos-acceptance -->

### 2b: Google Auth Platform (new UI)

Google has migrated the OAuth consent screen to a new "Google Auth Platform" interface. You may see one of two UI paths:

**New UI (Google Auth Platform)**:
- URL: [https://console.cloud.google.com/auth/overview](https://console.cloud.google.com/auth/overview)
- If you see "Get started" or "Configure consent screen", click it

**Legacy UI**:
- URL: [https://console.cloud.google.com/apis/credentials/consent](https://console.cloud.google.com/apis/credentials/consent)
- Follow the wizard steps

<!-- screenshot: google-auth-platform-overview -->

### 2c: Set User Type

| Your account type | User type to select | Notes |
|---|---|---|
| Google Workspace (company email) | **Internal** | Only users in your org can authenticate. No verification needed. |
| Personal Gmail (@gmail.com) | **External** | Any Google account can authenticate. Requires adding test users in Testing mode. |

> **Pitfall**: If you select "External" and leave the app in "Testing" mode, you MUST add yourself as a test user (see Step 2e). Otherwise you will get `403: access_denied` errors. Tokens in Testing mode also expire every 7 days.

<!-- screenshot: user-type-selection -->

### 2d: Fill in App Information

| Field | Value |
|-------|-------|
| App name | `gsuite-mcp` |
| User support email | Your email address |
| Developer contact email | Your email address |

You can leave all other fields blank (logo, app domain, authorized domains, etc.).

Click **Save and Continue**.

<!-- screenshot: app-information-form -->

### 2e: Add Test Users (External + Testing mode only)

If you selected "External" user type and your app is in "Testing" mode:

1. On the "Test users" section, click **Add users**
2. Enter your Google email address
3. Click **Save**

<!-- screenshot: add-test-users -->

> **Pitfall**: You must add the exact email you will authenticate with. If you skip this step, you will see `403: access_denied` with a message about "not completed verification".

### 2f: Publishing Status (Optional)

To avoid 7-day token expiration in Testing mode, you can publish the app:

1. Go to [Google Auth Platform Overview](https://console.cloud.google.com/auth/overview)
2. Click **Publish App** (or look for "Publishing status")
3. Confirm

This does NOT require Google verification for personal use with sensitive scopes. Your app will show an "unverified app" warning during first auth, which you can click through.

<!-- screenshot: publish-app-button -->

---

## Step 3: Enable the Required APIs

gsuite-mcp uses 10 Google APIs. Enable all of them in your project.

### Quick Method: Direct Links

Click each link to go directly to the API page and click **Enable**:

| # | API | Direct Enable Link |
|---|-----|--------------------|
| 1 | Gmail API | [Enable](https://console.cloud.google.com/apis/library/gmail.googleapis.com) |
| 2 | Google Calendar API | [Enable](https://console.cloud.google.com/apis/library/calendar-json.googleapis.com) |
| 3 | Google Drive API | [Enable](https://console.cloud.google.com/apis/library/drive.googleapis.com) |
| 4 | Google Docs API | [Enable](https://console.cloud.google.com/apis/library/docs.googleapis.com) |
| 5 | Google Sheets API | [Enable](https://console.cloud.google.com/apis/library/sheets.googleapis.com) |
| 6 | Google Slides API | [Enable](https://console.cloud.google.com/apis/library/slides.googleapis.com) |
| 7 | Google Tasks API | [Enable](https://console.cloud.google.com/apis/library/tasks.googleapis.com) |
| 8 | People API (Contacts) | [Enable](https://console.cloud.google.com/apis/library/people.googleapis.com) |
| 9 | Google Forms API | [Enable](https://console.cloud.google.com/apis/library/forms.googleapis.com) |
| 10 | Google Meet API | [Enable](https://console.cloud.google.com/apis/library/meet.googleapis.com) |

<!-- screenshot: api-enable-button -->

> **Pitfall**: Make sure you are in the correct project (check the project selector at top) before enabling APIs. If you enable them in a different project, your credentials will not work.

### Alternative: gcloud CLI

If you prefer the terminal:

```bash
# Set your project
gcloud config set project YOUR_PROJECT_ID

# Enable all APIs at once
gcloud services enable \
  gmail.googleapis.com \
  calendar-json.googleapis.com \
  drive.googleapis.com \
  docs.googleapis.com \
  sheets.googleapis.com \
  slides.googleapis.com \
  tasks.googleapis.com \
  people.googleapis.com \
  forms.googleapis.com \
  meet.googleapis.com
```

To check which APIs are currently enabled:

```bash
gcloud services list --enabled --filter="config.name:(gmail OR calendar OR drive OR docs OR sheets OR slides OR tasks OR people OR forms OR meet)"
```

### Verify with gsuite-mcp

After enabling APIs, you can verify with:

```bash
gsuite-mcp check
```

This checks that all required APIs are enabled and reports any that are missing with direct links to enable them.

---

## Step 4: Create OAuth Credentials

**URL**: [https://console.cloud.google.com/apis/credentials](https://console.cloud.google.com/apis/credentials)

1. Click **Create Credentials** at the top of the page
2. Select **OAuth client ID**

<!-- screenshot: create-credentials-dropdown -->

3. For **Application type**, select **Desktop app**

> **Pitfall**: You MUST select "Desktop app", not "Web application". Web application credentials require redirect URIs and will not work with gsuite-mcp's local OAuth flow. This is one of the most common mistakes.

<!-- screenshot: application-type-desktop -->

4. Name it `gsuite-mcp` (or anything you like)
5. Click **Create**

<!-- screenshot: oauth-client-created -->

### Download the Client Secret

After creation, you will see a dialog with your client ID and client secret.

**Download the JSON file immediately** by clicking the **Download JSON** button (download icon) in the dialog.

> **Pitfall**: If you dismiss this dialog, you can still download the JSON later from the Credentials page. Click the pencil/edit icon next to your OAuth client, then click "Download JSON" at the top. However, it is easiest to download it right away.

<!-- screenshot: download-client-secret -->

### Install the Client Secret

```bash
# Create the config directory
mkdir -p ~/.config/gsuite-mcp

# Move the downloaded file (adjust the filename if different)
mv ~/Downloads/client_secret_*.json ~/.config/gsuite-mcp/client_secret.json
```

> **Important**: The file MUST be named `client_secret.json` and placed in `~/.config/gsuite-mcp/`. gsuite-mcp will not find it under any other name or location (unless you use `--config-dir`).

---

## Step 5: Authenticate

```bash
gsuite-mcp auth
```

This will:
1. Open your default browser
2. Show the Google sign-in page
3. Ask you to grant permissions (you may see an "unverified app" warning -- click "Advanced" then "Go to gsuite-mcp")
4. Redirect back to the local server on success
5. Save your token to `~/.config/gsuite-mcp/credentials/<your-email>.json`

### Multiple Accounts

Run `gsuite-mcp auth` while signed into different Google accounts to add multiple accounts:

```bash
# First account
gsuite-mcp auth
# Sign in as alice@gmail.com

# Second account (open an incognito window or sign out first)
gsuite-mcp auth
# Sign in as bob@company.com
```

Then use the `account` parameter in tool calls:

```json
{"query": "is:unread", "account": "alice@gmail.com"}
```

---

## Step 6: Verify Everything

```bash
gsuite-mcp check
```

Expected output shows green checks for config, credentials, and all APIs:

```
Config:       OK (found client_secret.json)
Credentials:  OK (1 account authenticated)
APIs:         OK (all enabled)
```

If anything fails, the output includes specific fix instructions.

---

## Troubleshooting

### "access_denied" / "not completed verification"

**Cause**: Your app is in "Testing" mode and you are not listed as a test user.

**Fix**: Either:
- Add your email as a test user: [OAuth consent screen](https://console.cloud.google.com/apis/credentials/consent) > Test users > Add users
- Or publish the app: [Google Auth Platform](https://console.cloud.google.com/auth/overview) > Publish App

### "redirect_uri_mismatch"

**Cause**: You created a "Web application" credential instead of "Desktop app".

**Fix**: Delete the credential and create a new one with type "Desktop app". See [Step 4](#step-4-create-oauth-credentials).

### "invalid_client"

**Cause**: The `client_secret.json` file is corrupted, from a different project, or incorrectly named.

**Fix**:
1. Re-download from [Credentials page](https://console.cloud.google.com/apis/credentials)
2. Save as `~/.config/gsuite-mcp/client_secret.json`

### "SERVICE_DISABLED" / API not enabled

**Cause**: One or more APIs are not enabled in the GCP project.

**Fix**: Run `gsuite-mcp check` to see which APIs are missing, then enable them via the links in [Step 3](#step-3-enable-the-required-apis).

### "invalid_grant" / Token expired

**Cause**: Token expired (7 days in Testing mode) or was revoked.

**Fix**:
```bash
gsuite-mcp auth
```

Or delete the stale token and re-authenticate:
```bash
rm ~/.config/gsuite-mcp/credentials/<your-email>.json
gsuite-mcp auth
```

### Port already in use

**Cause**: The OAuth callback port (default 38917) is in use by another process.

**Fix**: Set a different port:
```bash
# Via environment variable
GSUITE_MCP_OAUTH_PORT=9000 gsuite-mcp auth

# Or in config.json
echo '{"oauth_port": 9000}' > ~/.config/gsuite-mcp/config.json
```

### "Unverified app" warning during auth

**Expected behavior** for unpublished apps. Click **Advanced** > **Go to gsuite-mcp (unsafe)** to continue. This warning goes away once you publish the app.

### Wrong GCP project

**Symptoms**: APIs are enabled but authentication fails, or vice versa.

**Fix**: Verify all resources are in the same project:
1. Check the project selector at the top of every GCP Console page
2. The client ID in `client_secret.json` starts with your project number (visible at [project settings](https://console.cloud.google.com/iam-admin/settings))

---

## gcloud CLI Reference

For users who prefer the command line over the Console UI.

### Create project

```bash
gcloud projects create gsuite-mcp-project --name="gsuite-mcp"
gcloud config set project gsuite-mcp-project
```

### Enable APIs

```bash
gcloud services enable \
  gmail.googleapis.com \
  calendar-json.googleapis.com \
  drive.googleapis.com \
  docs.googleapis.com \
  sheets.googleapis.com \
  slides.googleapis.com \
  tasks.googleapis.com \
  people.googleapis.com \
  forms.googleapis.com \
  meet.googleapis.com
```

### Configure OAuth consent screen

The consent screen cannot be fully configured via gcloud CLI. You must use the Console UI for:
- Setting user type (Internal/External)
- Adding test users
- Publishing the app

### Create OAuth credentials

```bash
gcloud auth application-default login --scopes=https://www.googleapis.com/auth/gmail.modify
```

> **Note**: gcloud CLI creates application default credentials, not OAuth Desktop app credentials. For gsuite-mcp, you need to create Desktop app credentials via the Console UI and download the `client_secret.json`. The gcloud CLI cannot create Desktop app OAuth credentials.

### List enabled APIs

```bash
gcloud services list --enabled
```

### Check project info

```bash
gcloud config get-value project
gcloud projects describe $(gcloud config get-value project)
```
