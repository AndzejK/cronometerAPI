# 🥑 Cronometer Daily Exporter

A simple Go script that exports your **Cronometer daily reports** (servings, biometrics, notes, etc.) to CSV and TXT files — either for a specific date or automatically for yesterday.

---

## 🧠 Overview

This tool helps you fetch and organize your daily Cronometer data into neatly formatted files.  
You can specify a date or let it automatically process **yesterday’s data**.

---

## 🚀 Usage

### 1. Run for a Specific Date

```bash
go run main.go 2025-10-30

This will generate files like:
servings_2025-10-30.csv
biometrics_2025-10-30.csv
notes_2025-10-30.csv
servings_formatted_2025-10-30.txt

### 2. Run Without a Date (Uses Yesterday)

go run main.go

If no date is provided, the script automatically uses yesterday’s date.

### 3. Invalid Date Example

go run main.go 10/30/2025

Output:
go run main.go 10/30/2025


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