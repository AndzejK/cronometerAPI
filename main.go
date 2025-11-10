package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/jrmycanady/gocronometer"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// --- Google Sheets Authentication Functions ---

func getClient(config *oauth2.Config) *http.Client {
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the " +
		"authorization code: \n%v\n", authURL)
	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}
	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func getSheetsService() (*sheets.Service, error) {
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		return nil, fmt.Errorf("unable to read client secret file: %v", err)
	}
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)
	srv, err := sheets.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Sheets client: %v", err)
	}
	return srv, nil
}

// --- Main Application Logic ---

func main() {
	// --- Configuration and Setup ---
	ctx := context.Background()
	cronometerClient := gocronometer.NewClient(nil)

	email := os.Getenv("CRONOMETER_EMAIL")
	password := os.Getenv("CRONOMETER_PASSWORD")
	reportsDir := os.Getenv("CRONOMETER_REPORTS_DIR")
	spreadsheetId := os.Getenv("SPREADSHEET_ID")
	sheetName := os.Getenv("GOOGLE_SHEET_NAME")

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

	// --- Action 2: Append Servings Totals to Google Sheets ---
	if spreadsheetId == "" || sheetName == "" {
		fmt.Println("\nSPREADSHEET_ID or GOOGLE_SHEET_NAME not set. Skipping Google Sheets upload.")
		return
	}

	fmt.Printf("\nPreparing to append servings totals to Google Sheet: %s\n", sheetName)

	// Read the servings CSV file we just saved
	file, err := os.Open(servingsCSVPath)
	if err != nil {
		log.Fatalf("Failed to open servings CSV for reading: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Failed to parse servings CSV: %v", err)
	}

	if len(records) <= 1 { // If 0 records or only header, no data rows to append
		fmt.Println("No data rows found in servings report (or only header). Nothing to append.")
		return
	}

	// Use all records *except* the header row
	var values [][]interface{}
	for _, record := range records[1:] { // Start from the second row
		row := make([]interface{}, len(record))
		for i, v := range record {
			row[i] = v
		}
		values = append(values, row)
	}

	// Get Sheets Service
	sheetsService, err := getSheetsService()
	if err != nil {
		log.Fatalf("Unable to initialize Google Sheets service: %v", err)
	}
	fmt.Println("✓ Successfully connected to Google Sheets.")

	// Append the data
	appendRange := sheetName // The sheet to append to
	valueRange := &sheets.ValueRange{
		Values: values,
	}
	_, err = sheetsService.Spreadsheets.Values.Append(spreadsheetId, appendRange, valueRange).ValueInputOption("USER_ENTERED").InsertDataOption("INSERT_ROWS").Do()
	if err != nil {
		log.Fatalf("Unable to append data to sheet: %v", err)
	}

	fmt.Printf("✓ Successfully appended servings data to sheet '%s'.\n", sheetName)

	// --- Action 3: Append Biometrics to Google Sheets ---
	biometricsSheetName := "biometricsReport" // As requested
	fmt.Printf("\nPreparing to append biometrics data to Google Sheet: %s\n", biometricsSheetName)

	// Read the biometrics CSV file we saved earlier
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
		// Use all records *except* the header row
		var biometricsValues [][]interface{}
		for _, record := range biometricsRecords[1:] { // Start from the second row
			row := make([]interface{}, len(record))
			for i, v := range record {
				row[i] = v
			}
			biometricsValues = append(biometricsValues, row)
		}

		// Append the data
		biometricsAppendRange := biometricsSheetName
		biometricsValueRange := &sheets.ValueRange{
			Values: biometricsValues,
		}
		_, err = sheetsService.Spreadsheets.Values.Append(spreadsheetId, biometricsAppendRange, biometricsValueRange).ValueInputOption("USER_ENTERED").InsertDataOption("INSERT_ROWS").Do()
		if err != nil {
			log.Fatalf("Unable to append biometrics data to sheet: %v", err)
		}
		fmt.Printf("✓ Successfully appended biometrics data to sheet '%s'.\n", biometricsSheetName)
	}

	fmt.Println("\n=== All tasks complete! ===")
}