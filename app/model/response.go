package model

// Meta adalah metadata pagination untuk endpoint daftar (offset-based).
// Sesuai format yang diminta soal: current_page, per_page, total, last_page.
type Meta struct {
	CurrentPage int `json:"current_page"`
	PerPage     int `json:"per_page"`
	Total       int `json:"total"`
	LastPage    int `json:"last_page"`
}

// NewMeta membuat objek Meta dari parameter pagination.
// LastPage dihitung otomatis dari total dan perPage.
func NewMeta(currentPage, perPage, total int) Meta {
	lastPage := total / perPage
	if total%perPage != 0 {
		lastPage++
	}
	if lastPage < 1 {
		lastPage = 1
	}
	return Meta{
		CurrentPage: currentPage,
		PerPage:     perPage,
		Total:       total,
		LastPage:    lastPage,
	}
}
