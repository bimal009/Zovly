// internal/repository/helpers.go
package repository

import "encoding/json"

// orEmptySlice returns an empty slice if input is nil.
// Needed so pq.Array never stores NULL for image columns.
func orEmptySlice(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

// orNullJSON returns nil (→ SQL NULL) for empty/absent JSON, so an empty
// json.RawMessage never gets written as invalid jsonb.
func orNullJSON(raw json.RawMessage) any {
	if len(raw) == 0 {
		return nil
	}
	return []byte(raw)
}
