# Cronometer to Google Sheets Exporter

This Go application automates exporting your daily data from Cronometer. It performs two main actions:

1.  Saves all daily reports (servings, biometrics, notes) as CSV files to a local directory on your computer.
2.  Appends all data rows from the daily `servings` and `biometrics` reports directly to specified Google Sheets.

---

## Features

- **Dual Export:** Saves raw data locally and pushes key data to the cloud.
- **Google Sheets Integration:** Automatically appends daily servings and biometrics data to separate sheets.
- **Flexible Date Selection:** Run for a specific date or let it default to yesterday.
- **Secure and Customizable:** Uses environment variables for all credentials and configuration.

---

## ⚙️ Setup

Before you can run the application, you need to set up your Cronometer and Google credentials.

### 1. Google API Credentials (One-Time Setup)

This application uses the Google Sheets API. You must create your own credentials to allow the application to access your sheets.

1.  **Go to the Google Cloud Console:** [https://console.cloud.google.com/](https://console.cloud.google.com/)
2.  **Create a New Project:** Give it a name like `Cronometer-Sheets-Integration`.
3.  **Enable the Google Sheets API:** In the search bar, find and **ENABLE** the "Google Sheets API".
4.  **Configure OAuth Consent Screen:**
    - Go to **APIs & Services > OAuth consent screen**.
    - Choose **"External"** and click **Create**.
    - Fill in the required fields (App name, your email). Click **Save and Continue** through the "Scopes" and "Optional Info" pages.
    - On the **Test users** page, add your own Google email address. This is a critical step.
5.  **Create Credentials:**
    - Go to **APIs & Services > Credentials**.
    - Click **+ CREATE CREDENTIALS** and select **"OAuth client ID"**.
    - Set the **Application type** to **"Desktop app"** and click **Create**.
6.  **Download the File:**
    - In the credentials list, find the one you just created and click the **download icon**.
    - **IMPORTANT:** Rename the downloaded file to `credentials.json` and place it in the root of this project directory. This file should be ignored by Git and never be made public.

### 2. Set Environment Variables

This application is configured using environment variables. Choose the section for your operating system.

**On macOS/Linux (bash/zsh):**

```bash
# Cronometer Credentials
export CRONOMETER_EMAIL="your_email@example.com"
export CRONOMETER_PASSWORD="your_secure_password"

# Google Sheets Configuration
export SPREADSHEET_ID="your_google_sheet_id_here"
export GOOGLE_SHEET_NAME="name_of_sheet_for_servings"   # Target sheet for servings data
export BIOMETRICS_SHEET_NAME="name_of_sheet_for_biometrics"  # Optional, defaults to "biometricsReport"
# The sheet for biometrics is currently set to "biometricsReport" in the code.

# Optional: Local reports directory
# export CRONOMETER_REPORTS_DIR="/path/to/your/reports"
```

**On Windows (PowerShell):**

```powershell
# Cronometer Credentials
$env:CRONOMETER_EMAIL="your_email@example.com"
$env:CRONOMETER_PASSWORD="your_secure_password"

# Google Sheets Configuration
$env:SPREADSHEET_ID="your_google_sheet_id_here"
$env:GOOGLE_SHEET_NAME="name_of_sheet_for_servings"   # Target sheet for servings data
$env:BIOMETRICS_SHEET_NAME="name_of_sheet_for_biometrics"  # Optional, defaults to "biometricsReport"

# Optional: Local reports directory
# $env:CRONOMETER_REPORTS_DIR="C:\path\to\your\reports"
```

**On Windows (Command Prompt - cmd.exe):**

```cmd
set CRONOMETER_EMAIL="your_email@example.com"
set CRONOMETER_PASSWORD="your_secure_password"
set SPREADSHEET_ID="your_google_sheet_id_here"
set GOOGLE_SHEET_NAME="name_of_sheet_for_servings"
set BIOMETRICS_SHEET_NAME="name_of_sheet_for_biometrics"

:: Optional: Local reports directory
:: set CRONOMETER_REPORTS_DIR="C:\path\to\your\reports"
```

---

## 🚀 Running the Application

1.  **Build the application:**
    ```bash
    go build .
    ```
2.  **Run for yesterday's data:**
    ```bash
    ./coffeeproject
    ```
3.  **Run for a specific date (YYYY-MM-DD):**
    ```bash
    ./coffeeproject 2025-11-09
    ```

### First-Time Google Authentication

The very first time you run the application, it will prompt you to authorize it with Google:

1. **The application will print a URL in the terminal.** Copy and paste this entire URL into your browser.

2. **Choose your Google account** and grant the requested permissions (Google Sheets access).

3. **After authorization, Google will redirect you to a URL that looks like this:**
```
   http://localhost/?state=state-token&code=4/0Ab32j9322yqmpk6vj05jmT4mXzoZ-N5C7DemoTesTKHiUOoBPi2Iz3xYu6SnOWh21YlRg&scope=https://www.googleapis.com/auth/spreadsheets
```

4. **Copy ONLY the authorization code** - the part that comes after `code=` and stops **before** the `&` symbol:
   
   **✅ CORRECT - Copy this (no `&` at the end):**
```
   4/0Ab32j9322yqmpk6vj05jmT4mXzoZ-N5C7DemoTesTKHiUOoBPi2Iz3xYu6SnOWh21YlRg
```
   
   **❌ WRONG - Do NOT include the `&` or anything after it:**
```
   4/0Ab32j9322yqmpk6vj05jmT4mXzoZ-N5C7DemoTesTKHiUOoBPi2Iz3xYu6SnOWh21YlRg&scope=https://www.googleapis.com/auth/spreadsheets
```
   
   **Visual guide:**
```
   http://localhost/?state=state-token&code=4/0Ab32j9322yqmpk6vj05jmT4mXzoZ-N5C7DemoTesTKHiUOoBPi2Iz3xYu6SnOWh21YlRg&scope=https://www.googleapis.com/auth/spreadsheets
                                          ↑                                                                              ↑
                                      START copying here                                                          STOP before the &
```

5. **Paste ONLY this code** (without the `&`) back into your terminal where the application is waiting.

6. Press Enter, and the application will complete the authentication.

A `token.json` file will be created to store your login session, so you won't have to do this again unless you delete it or the token expires.

**Common Mistakes:**
- ❌ Including `&scope=...` at the end
- ❌ Copying the entire URL
- ❌ Including the `&` symbol
- ✅ Only copy the code value itself (starts with `4/0` in the example)

**Troubleshooting:**
- If the browser shows "This site can't be reached" - that's normal! Just copy the code from the URL bar
- If authentication fails, double-check you didn't include the `&` or anything after it
- The code should typically start with `4/0` and contain only letters, numbers, hyphens, and underscores