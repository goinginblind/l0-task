package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestValidation(t *testing.T) {
	testCases := []struct {
		name        string
		filepath    string
		expectValid bool
	}{
		{
			name:        "Valid data",
			filepath:    "testdata/valn_ok.json",
			expectValid: true,
		},
		{
			name:        "Invalid data",
			filepath:    "testdata/valn_non1.json",
			expectValid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := os.ReadFile(tc.filepath)
			if err != nil {
				t.Fatalf("Failed to read test file: %v", err)
			}

			var order Order
			if err := json.Unmarshal(data, &order); err != nil {
				// For this test, we assume JSON is valid and focus on validation logic
				// but if it's not, it's a test setup problem (so a 'you' problem)
				if tc.expectValid {
					t.Fatalf("Failed to unmarshal what should be valid json: %v", err)
				}
				// If we expect invalid, an unmarshal error on a malformed struct is a pass.
				return
			}

			ok, err := order.Validate()

			if tc.expectValid {
				if !ok {
					t.Errorf("Expected order to be valid, but it was not. Error: %v", err)
				}
				if err != nil {
					t.Errorf("Expected no error, but got: %v", err)
				}
			} else {
				if ok {
					t.Error("Expected order to be invalid, but it was valid.")
				}
				if err == nil {
					t.Error("Expected an error, but got none.")
				}
			}
		})
	}
}
