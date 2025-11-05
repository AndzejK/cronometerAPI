package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/jrmycanady/gocronometer"
)

func main() {
	ctx := context.Background()
	client := gocronometer.NewClient(nil)

	// Get credentials from environment variables
	email := os.Getenv("CRONOMETER_EMAIL")
	password := os.Getenv("CRONOMETER_PASSWORD")

	// Check if credentials are set
	if email == "" || password == "" {
		log.Fatalf("Error: CRONOMETER_EMAIL and CRONOMETER_PASSWORD environment variables must be set")
	}

	// Login with environment variables
	if err := client.Login(ctx, email, password); err != nil {
		log.Fatalf("login failed: %v", err)
	}
	defer client.Logout(ctx)

	// Get reports directory from environment variable, or use default
	reportsDir := os.Getenv("CRONOMETER_REPORTS_DIR")
	if reportsDir == "" {
		reportsDir = `C:\GO\reports`
	}

	// Create the reports directory if it doesn't exist
	if err := os.MkdirAll(reportsDir, 0755); err != nil {
		log.Fatalf("failed to create reports directory: %v", err)
	}

	// Get date from command line argument
	var dateStr string
	if len(os.Args) < 2 {
		// If no date provided, use yesterday's date
		dateStr = time.Now().AddDate(0, 0, -1).Format("2006-01-02")
		fmt.Printf("No date provided, using yesterday: %s\n", dateStr)
	} else {
		dateStr = os.Args[1]
	}

	// Parse the date string
	targetDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		log.Fatalf("Invalid date format. Please use YYYY-MM-DD format (e.g., 2025-10-30): %v", err)
	}

	// Set start and end to the same day
	start := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 1) // End is start of next day

	fmt.Printf("Exporting data for: %s\n", dateStr)
	fmt.Printf("Saving files to: %s\n\n", reportsDir)

	// Export Servings (food/drinks)
	fmt.Println("Exporting servings...")
	servings, err := client.ExportServings(ctx, start, end)
	if err != nil {
		log.Fatalf("export servings: %v", err)
	}
	filename := filepath.Join(reportsDir, fmt.Sprintf("servings_%s.csv", dateStr))
	err = os.WriteFile(filename, []byte(servings), 0644)
	if err != nil {
		log.Fatalf("failed to write servings CSV: %v", err)
	}
	fmt.Printf("✓ Servings saved to %s\n", filename)

	// Export Biometrics (body measurements)
	fmt.Println("Exporting biometrics...")
	biometrics, err := client.ExportBiometrics(ctx, start, end)
	if err != nil {
		log.Fatalf("export biometrics: %v", err)
	}
	filename = filepath.Join(reportsDir, fmt.Sprintf("biometrics_%s.csv", dateStr))
	err = os.WriteFile(filename, []byte(biometrics), 0644)
	if err != nil {
		log.Fatalf("failed to write biometrics CSV: %v", err)
	}
	fmt.Printf("✓ Biometrics saved to %s\n", filename)

	// Export Notes (diary notes)
	fmt.Println("Exporting notes...")
	notes, err := client.ExportNotes(ctx, start, end)
	if err != nil {
		log.Fatalf("export notes: %v", err)
	}
	filename = filepath.Join(reportsDir, fmt.Sprintf("notes_%s.csv", dateStr))
	err = os.WriteFile(filename, []byte(notes), 0644)
	if err != nil {
		log.Fatalf("failed to write notes CSV: %v", err)
	}
	fmt.Printf("✓ Notes saved to %s\n", filename)

	// Create formatted servings output
	fmt.Println("Creating formatted servings output...")
	records, err := client.ExportServingsParsedWithLocation(ctx, start, end, time.Local)
	if err != nil {
		log.Fatalf("parse servings: %v", err)
	}

	filename = filepath.Join(reportsDir, fmt.Sprintf("servings_formatted_%s.txt", dateStr))
	file, err := os.Create(filename)
	if err != nil {
		log.Fatalf("failed to create formatted file: %v", err)
	}
	defer file.Close()

	for _, r := range records {
		line := fmt.Sprintf("%s %s ate %.1f %s (%.0f kcal)\n",
			r.RecordedTime.Format(time.RFC3339), r.Group, r.QuantityValue, r.QuantityUnits, r.EnergyKcal)
		file.WriteString(line)
	}
	fmt.Printf("✓ Formatted servings saved to %s\n", filename)

	fmt.Println("\n=== All exports complete! ===")
	fmt.Printf("All files saved to: %s\n", reportsDir)
	fmt.Println("Files created:")
	fmt.Printf("  - servings_%s.csv\n", dateStr)
	fmt.Printf("  - biometrics_%s.csv\n", dateStr)
	fmt.Printf("  - notes_%s.csv\n", dateStr)
	fmt.Printf("  - servings_formatted_%s.txt\n", dateStr)
}
