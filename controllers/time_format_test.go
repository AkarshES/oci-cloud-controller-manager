package controllers

import (
	"regexp"
	"testing"
)

func TestTimeFormat(t *testing.T) {

	// Get the current time and format it
	timestamp := CreateRepairTaint().Value
	t.Logf("Timestamp: %s", timestamp)

	// --- Verification ---

	// 1. Verify the format using a regular expression (more robust)
	// The pattern checks for YYYY-MM-DD-HH-MM-SS format
	regexPattern := `^\d{4}-\d{2}-\d{2}-\d{2}-\d{2}-\d{2}$`
	matched, err := regexp.MatchString(regexPattern, timestamp)
	if err != nil || !matched {
		t.Errorf("Timestamp format is incorrect. Got: %s", timestamp)
	}

	// --- Output to console ---
	t.Logf("Timestamp verified: %s", timestamp)
}
