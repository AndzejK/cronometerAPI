# Cronometer to Google Sheets Exporter

This Go application automates exporting your daily data from Cronometer. It performs two main actions:
1.  Saves all daily reports (servings, biometrics, notes) as CSV files to a local directory on your computer.
2.  Appends all data rows from the daily `servings` report directly to a Google Sheet of your choice.

---

## Features

-   **Dual Export:** Saves raw data locally and pushes key data to the cloud.
-   **Google Sheets Integration:** Automatically appends daily food diary entries to a specified Google Sheet.
-   **Flexible Date Selection:** Run for a specific date or let it default to yesterday.
-   **Secure Configuration:** Uses environment variables for all credentials and configuration, so no secrets are stored in the code.

---

## ⚙️ Setup

Before you can run the application, you need to set up your Cronometer and Google credentials.

### 1. Google API Credentials (One-Time Setup)

This application uses the Google Sheets API. You must create your own credentials to allow the application to access your sheets.

1.  **Go to the Google Cloud Console:** [https://console.cloud.google.com/](https://console.cloud.google.com/)
2.  **Create a New Project:** Give it a name like `Cronometer-Sheets-Integration`.
3.  **Enable the Google Sheets API:** In the search bar, find and **ENABLE** the "Google Sheets API".
4.  **Configure OAuth Consent Screen:**
    *   Go to **APIs & Services > OAuth consent screen**.
    *   Choose **"External"** and click **Create**.
    *   Fill in the required fields (App name, your email). Click **Save and Continue** through the "Scopes" and "Optional Info" pages.
    *   On the **Test users** page, add your own Google email address. This is a critical step.
5.  **Create Credentials:**
    *   Go to **APIs & Services > Credentials**.
    *   Click **+ CREATE CREDENTIALS** and select **"OAuth client ID"**.
    *   Set the **Application type** to **"Desktop app"** and click **Create**.
6.  **Download the File:**
    *   In the credentials list, find the one you just created and click the **download icon**.
    *   **IMPORTANT:** Rename the downloaded file to `credentials.json` and place it in the root of this project directory. This file should be ignored by Git and never be made public.

### 2. Set Environment Variables

This application is configured using environment variables.

**On macOS/Linux:**
```bash
# Cronometer Credentials
export CRONOMETER_EMAIL="your_email@example.com"
export CRONOMETER_PASSWORD="your_secure_password"

# Google Sheets Configuration
export SPREADSHEET_ID="your_google_sheet_id_here"
export GOOGLE_SHEET_NAME="reportCronoNov" # Or whatever your sheet is named

# Optional: Local reports directory (defaults to a folder in your home dir)
# export CRONOMETER_REPORTS_DIR="/path/to/your/reports"
```

**On Windows (Command Prompt):**
```cmd
set CRONOMETER_EMAIL="your_email@example.com"
set CRONOMETER_PASSWORD="your_secure_password"
set SPREADSHEET_ID="your_google_sheet_id_here"
set GOOGLE_SHEET_NAME="reportCronoNov"
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
1.  It will print a URL in the terminal. Copy and paste this into your browser.
2.  Choose your Google account and grant the requested permissions.
3.  Google will give you an authorization code. Copy this code.
4.  Paste the code back into your terminal where the application is waiting.

<<<<<<< HEAD
Invalid date format. Please use YYYY-MM-DD format
✅ Correct format: YYYY-MM-DD
❌ Wrong format: MM/DD/YYYY

### 📅 Bonus: Export a Date Range (Optional Feature)

If you want to export multiple days at once, you can extend the script to accept a date range:

go run main.go 2025-10-01 2025-10-31

### 📂 Custom Output Directory

You can change where reports are saved without modifying the code — simply use an environment variable.

Example Go snippet:
// Get reports directory from environment variable, or use default
reportsDir := os.Getenv("CRONOMETER_REPORTS_DIR")
if reportsDir == "" {
	reportsDir = `C:\GO\reports`
}

### Set environment variable:
``` PowerShell (Windows)
$env:CRONOMETER_REPORTS_DIR = "D:\MyReports"

``` Linux / macOS (bash)
export CRONOMETER_REPORTS_DIR=/home/user/reports

### 🔐 Credentials

Before running, make sure you’ve set your Cronometer login details as environment variables:

export CRONOMETER_EMAIL="your@email.com"
export CRONOMETER_PASSWORD="your_password"


⚠️ Security tip: Avoid hardcoding your credentials in code.
Use environment variables or a .env file (excluded from Git) for safety.

### 🧩 Example Folder Structure
📁 reports/
 ┣ 📄 servings_2025-10-30.csv
 ┣ 📄 biometrics_2025-10-30.csv
 ┣ 📄 notes_2025-10-30.csv
 ┗ 📄 servings_formatted_2025-10-30.txt
=======
A `token.json` file will be created to store your login session, so you won't have to do this again unless you delete it.
>>>>>>> 0c30ce0a4104f23b58960bc0ef8220e0e962afde
