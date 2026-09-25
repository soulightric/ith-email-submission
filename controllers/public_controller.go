package controllers

import (
	"database/sql"
	"html/template"
	"net/http"
	"net/url"
	"strconv"

	"monitoring-email-ith/models"
)

// PublicController menangani halaman yang bisa diakses tanpa login.
type PublicController struct {
	DB   *sql.DB
	Tmpl *template.Template
}

func NewPublicController(db *sql.DB, tmpl *template.Template) *PublicController {
	return &PublicController{DB: db, Tmpl: tmpl}
}

// Index menampilkan tabel status usulan: No, Jenis Usulan, Nama, NIP/NIM,
// Prodi/Unit, Status (detail muncul lewat modal saat diklik). Dibatasi 100
// baris terbaru, bisa dicari berdasarkan nama atau NIM.
func (c *PublicController) Index(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("q")
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = 1
	}
	const pageSize = 100

	requests, err := models.ListPublicEmailRequests(c.DB, search, page, pageSize)
	if err != nil {
		http.Error(w, "Gagal memuat data, silakan coba lagi", http.StatusInternalServerError)
		return
	}
	filteredTotal, err := models.CountPublicEmailRequests(c.DB, search)
	if err != nil {
		http.Error(w, "Gagal menghitung data, silakan coba lagi", http.StatusInternalServerError)
		return
	}
	totalPages := (filteredTotal + pageSize - 1) / pageSize
	if totalPages > 0 && page > totalPages {
		page = totalPages
		requests, err = models.ListPublicEmailRequests(c.DB, search, page, pageSize)
		if err != nil {
			http.Error(w, "Gagal memuat halaman, silakan coba lagi", http.StatusInternalServerError)
			return
		}
	}
	pageURL := func(number int) string {
		values := url.Values{}
		if search != "" {
			values.Set("q", search)
		}
		values.Set("page", strconv.Itoa(number))
		return "/?" + values.Encode()
	}
	pageLinks := make([]map[string]interface{}, 0, totalPages)
	for number := 1; number <= totalPages; number++ {
		pageLinks = append(pageLinks, map[string]interface{}{"Number": number, "URL": pageURL(number), "Current": number == page})
	}

	data := map[string]interface{}{
		"Requests":      requests,
		"Search":        search,
		"FilteredTotal": filteredTotal,
		"Page":          page,
		"PageOffset":    (page - 1) * pageSize,
		"TotalPages":    totalPages,
		"HasPrevious":   page > 1,
		"PreviousURL":   pageURL(page - 1),
		"HasNext":       page < totalPages,
		"NextURL":       pageURL(page + 1),
		"PageLinks":     pageLinks,
	}
	if err := c.Tmpl.ExecuteTemplate(w, "public_index.html", data); err != nil {
		http.Error(w, "Gagal menampilkan halaman", http.StatusInternalServerError)
	}
}
