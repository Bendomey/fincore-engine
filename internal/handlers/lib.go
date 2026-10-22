package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Bendomey/fincore-engine/internal/lib"
)

func getPopulateFields(r *http.Request) *[]string {
	var populateFields *[]string = nil
	populate := r.URL.Query().Get("populate")

	if populate != "" {
		fields := strings.Split(populate, ",")
		populateFields = &fields
	}

	return populateFields
}

// writeError writes the standard single-message error envelope.
func writeError(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]any{
		"errors": map[string]string{
			"message": message,
		},
	})
}

var accountListOptions = lib.ListOptions{
	OrderBy:      []string{"created_at", "updated_at", "code", "name", "type"},
	SearchFields: []string{"code", "name", "description"},
	Populate:     map[string]string{"ParentAccount": "ParentAccount"},
}

var journalEntryListOptions = lib.ListOptions{
	OrderBy:      []string{"created_at", "updated_at", "transaction_date", "posted_at", "reference"},
	SearchFields: []string{"reference"},
	Populate: map[string]string{
		"JournalEntryLines": "JournalEntryLines",
		"Account":           "JournalEntryLines.Account",
	},
}
