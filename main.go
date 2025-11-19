package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jrmycanady/gocronometer"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

const (
	tokenFile       = "token.json"
	credentialsFile = "credentials.json"
)

func main() {
	// --- Configuration and Setup ---
	ctx := context.Background()
	cronometerClient := gocronometer.NewClient(nil)

	email := os.Getenv("CRONOMETER_EMAIL")
	password := os.Getenv("CRONOMETER_PASSWORD")
	reportsDir := os.Getenv("CRONOMETER_REPORTS_DIR")
	spreadsheetId := os.Getenv("SPREADSHEET_ID")
	sheetName := os.Getenv("GOOGLE_SHEET_NAME")
	biometricsSheetName := os.Getenv("BIOMETRICS_SHEET_NAME")

	if email == "" || password == "" {
		log.Fatalf("Error: CRONOMETER_EMAIL and CRONOMETER_PASSWORD environment variables must be set")
	}
	if reportsDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("failed to get user home directory: %v", err)
		}
		reportsDir = filepath.Join(homeDir, "cronometer_reports")
	}
	if biometricsSheetName == "" {
		biometricsSheetName = "biometricsReport" // Default value
	}

	// --- Login to Cronometer ---
	if err := cronometerClient.Login(ctx, email, password); err != nil {
		log.Fatalf("Cronometer login failed: %v", err)
	}
	defer cronometerClient.Logout(ctx)
	fmt.Println("✓ Successfully logged into Cronometer.")

	if err := os.MkdirAll(reportsDir, 0755); err != nil {
		log.Fatalf("Failed to create reports directory: %v", err)
	}

	// --- Get Date ---
	var dateStr string
	if len(os.Args) < 2 {
		dateStr = time.Now().AddDate(0, 0, -1).Format("2006-01-02")
		fmt.Printf("No date provided, using yesterday: %s\n", dateStr)
	} else {
		dateStr = os.Args[1]
	}
	targetDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		log.Fatalf("Invalid date format. Please use YYYY-MM-DD format (e.g., 2025-10-30): %v", err)
	}
	start := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 1)

	fmt.Printf("\nExporting local reports for: %s\n", dateStr)
	fmt.Printf("Saving files to: %s\n\n", reportsDir)

	// --- Action 1: Save All Reports Locally ---
	servingsCSVPath := filepath.Join(reportsDir, fmt.Sprintf("servings_%s.csv", dateStr))

	// Export Servings (food/drinks)
	servingsData, err := cronometerClient.ExportServings(ctx, start, end)
	if err != nil {
		log.Fatalf("export servings: %v", err)
	}
	if err = os.WriteFile(servingsCSVPath, []byte(servingsData), 0644); err != nil {
		log.Fatalf("failed to write servings CSV: %v", err)
	}
	fmt.Printf("✓ Servings report saved to %s\n", servingsCSVPath)

	// Export Biometrics
	biometricsData, err := cronometerClient.ExportBiometrics(ctx, start, end)
	if err != nil {
		log.Fatalf("export biometrics: %v", err)
	}
	biometricsPath := filepath.Join(reportsDir, fmt.Sprintf("biometrics_%s.csv", dateStr))
	if err = os.WriteFile(biometricsPath, []byte(biometricsData), 0644); err != nil {
		log.Fatalf("failed to write biometrics CSV: %v", err)
	}
	fmt.Printf("✓ Biometrics report saved to %s\n", biometricsPath)

	// Export Notes
	notesData, err := cronometerClient.ExportNotes(ctx, start, end)
	if err != nil {
		log.Fatalf("export notes: %v", err)
	}
	notesPath := filepath.Join(reportsDir, fmt.Sprintf("notes_%s.csv", dateStr))
	if err = os.WriteFile(notesPath, []byte(notesData), 0644); err != nil {
		log.Fatalf("failed to write notes CSV: %v", err)
	}
	fmt.Printf("✓ Notes report saved to %s\n", notesPath)
	fmt.Println("\n--- Local exports complete! ---")

	// --- Action 2: Append Servings to Google Sheets ---
	if spreadsheetId == "" || sheetName == "" {
		fmt.Println("\n⚠ SPREADSHEET_ID or GOOGLE_SHEET_NAME not set. Skipping Google Sheets upload.")
		fmt.Println("  Set these environment variables to enable Google Sheets integration.")
		return
	}

	fmt.Printf("\nPreparing to append servings data to Google Sheet: %s\n", sheetName)

	// Get Google Sheets service with token refresh handling
	sheetsService, err := getGoogleSheetsService(ctx)
	if err != nil {
		log.Fatalf("Unable to initialize Google Sheets service: %v", err)
	}
	fmt.Println("✓ Successfully connected to Google Sheets.")

	// Read and parse the servings CSV file
	servingsFile, err := os.Open(servingsCSVPath)
	if err != nil {
		log.Fatalf("Failed to open servings CSV for reading: %v", err)
	}
	defer servingsFile.Close()

	servingsReader := csv.NewReader(servingsFile)
	servingsRecords, err := servingsReader.ReadAll()
	if err != nil {
		log.Fatalf("Failed to parse servings CSV: %v", err)
	}

	if len(servingsRecords) <= 1 {
		fmt.Println("No data rows found in servings report (or only header). Nothing to append.")
	} else {
		// Convert CSV records to interface{} for Google Sheets API (skip header)
		var servingsValues [][]interface{}
		for _, record := range servingsRecords[1:] {
			row := make([]interface{}, len(record))
			for i, v := range record {
				row[i] = v
			}
			servingsValues = append(servingsValues, row)
		}

		// Append servings data to Google Sheets with retry logic
		if err := appendToSheetWithRetry(ctx, sheetsService, spreadsheetId, sheetName, servingsValues); err != nil {
			log.Fatalf("Unable to append servings data: %v", err)
		}
		fmt.Printf("✓ Successfully appended servings data to sheet '%s'.\n", sheetName)
	}

	// --- Action 3: Append Biometrics to Google Sheets ---
	fmt.Printf("\nPreparing to append biometrics data to Google Sheet: %s\n", biometricsSheetName)

	// Read and parse the biometrics CSV file
	biometricsFile, err := os.Open(biometricsPath)
	if err != nil {
		log.Fatalf("Failed to open biometrics CSV for reading: %v", err)
	}
	defer biometricsFile.Close()

	biometricsReader := csv.NewReader(biometricsFile)
	biometricsRecords, err := biometricsReader.ReadAll()
	if err != nil {
		log.Fatalf("Failed to parse biometrics CSV: %v", err)
	}

	if len(biometricsRecords) <= 1 {
		fmt.Println("No data rows found in biometrics report. Nothing to append.")
	} else {
		// Convert CSV records to interface{} for Google Sheets API (skip header)
		var biometricsValues [][]interface{}
		for _, record := range biometricsRecords[1:] {
			row := make([]interface{}, len(record))
			for i, v := range record {
				row[i] = v
			}
			biometricsValues = append(biometricsValues, row)
		}

		// Append biometrics data to Google Sheets with retry logic
		if err := appendToSheetWithRetry(ctx, sheetsService, spreadsheetId, biometricsSheetName, biometricsValues); err != nil {
			log.Fatalf("Unable to append biometrics data: %v", err)
		}
		fmt.Printf("✓ Successfully appended biometrics data to sheet '%s'.\n", biometricsSheetName)
	}

	fmt.Println("\n=== All tasks complete! ===")
}

// --- Google Sheets Service with Token Refresh ---

func getGoogleSheetsService(ctx context.Context) (*sheets.Service, error) {
	b, err := os.ReadFile(credentialsFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read credentials file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, sheets.SpreadsheetsScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse credentials: %v", err)
	}

	// Get token with refresh handling
	token, err := getTokenWithRefresh(ctx, config)
	if err != nil {
		return nil, err
	}

	// Create HTTP client with the token
	httpClient := config.Client(ctx, token)

	// Create Sheets service
	srv, err := sheets.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("unable to create Sheets service: %v", err)
	}

	return srv, nil
}

func getTokenWithRefresh(ctx context.Context, config *oauth2.Config) (*oauth2.Token, error) {
	token, err := tokenFromFile(tokenFile)
	if err != nil {
		// Token file doesn't exist or is invalid - need to re-authenticate
		return getTokenFromWeb(config)
	}

	// Check if token is expired
	if token.Expiry.Before(time.Now()) {
		fmt.Println("⚠ Token expired, refreshing...")

		// Try to refresh the token
		tokenSource := config.TokenSource(ctx, token)
		newToken, err := tokenSource.Token()
		if err != nil {
			// Refresh failed - need to re-authenticate
			fmt.Println("⚠ Token refresh failed. Re-authentication required.")
			return getTokenFromWeb(config)
		}

		// Save the new token
		if err := saveToken(tokenFile, newToken); err != nil {
			log.Printf("Warning: Failed to save refreshed token: %v", err)
		} else {
			fmt.Println("✓ Token refreshed successfully")
		}

		return newToken, nil
	}

	return token, nil
}

func getTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("\nGo to the following link in your browser:\n%v\n\n", authURL)
	fmt.Println("After authorizing, you'll be redirected to a URL like:")
	fmt.Println("http://localhost/?state=state-token&code=AUTHORIZATION_CODE&scope=...")
	fmt.Println("\nCopy ONLY the authorization code (the part between 'code=' and '&scope')")
	fmt.Print("\nEnter the authorization code: ")

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, fmt.Errorf("unable to read authorization code: %v", err)
	}

	token, err := config.Exchange(context.Background(), authCode)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve token: %v", err)
	}

	// Save the token for future use
	if err := saveToken(tokenFile, token); err != nil {
		log.Printf("Warning: Failed to save token: %v", err)
	}

	return token, nil
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	token := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(token)
	return token, err
}

func saveToken(path string, token *oauth2.Token) error {
	fmt.Printf("Saving token to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("unable to cache token: %v", err)
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(token)
}

func appendToSheetWithRetry(ctx context.Context, srv *sheets.Service, spreadsheetID, sheetName string, data [][]interface{}) error {
	valueRange := &sheets.ValueRange{
		Values: data,
	}

	_, err := srv.Spreadsheets.Values.Append(
		spreadsheetID,
		sheetName,
		valueRange,
	).ValueInputOption("USER_ENTERED").InsertDataOption("INSERT_ROWS").Context(ctx).Do()

	if err != nil {
		// Check if it's a token error
		if isTokenError(err) {
			fmt.Println("\n⚠ Token error detected. Please re-authenticate.")
			fmt.Println("  Solution: Delete token.json and run the program again")
			fmt.Printf("  Command: del %s (Windows) or rm %s (Mac/Linux)\n", tokenFile, tokenFile)
		}
		return fmt.Errorf("unable to append data: %v", err)
	}

	return nil
}

func isTokenError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "invalid_grant") ||
		strings.Contains(errStr, "token") && strings.Contains(errStr, "expired") ||
		strings.Contains(errStr, "token") && strings.Contains(errStr, "revoked")
}
