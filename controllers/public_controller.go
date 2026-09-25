package controllers

import (
	"database/sql"
	"html/template"
	"net/http"

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

	requests, err := models.ListPublicEmailRequests(c.DB, search)
	if err != nil {
		http.Error(w, "Gagal memuat data, silakan coba lagi", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Requests": requests,
		"Search":   search,
	}
	if err := c.Tmpl.ExecuteTemplate(w, "public_index.html", data); err != nil {
		http.Error(w, "Gagal menampilkan halaman", http.StatusInternalServerError)
	}
}
