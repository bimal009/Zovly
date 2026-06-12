// internal/repository/helpers.go
package repository

// orEmptySlice returns an empty slice if input is nil.
// Needed so pq.Array never stores NULL for image columns.
func orEmptySlice(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
