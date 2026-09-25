package controllers

import (
	"database/sql"
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"monitoring-email-ith/middleware"
	"monitoring-email-ith/models"
)

// AdminController menangani dashboard yang butuh login (dilindungi
// middleware.RequireAuth di routes).
type AdminController struct {
	DB   *sql.DB
	Tmpl *template.Template
}

func NewAdminController(db *sql.DB, tmpl *template.Template) *AdminController {
	return &AdminController{DB: db, Tmpl: tmpl}
}

// Dashboard mirip tabel public, ditambah kolom Aksi dan filter berdasarkan
// jenis usulan, rentang tanggal, dan status.
func (c *AdminController) Dashboard(w http.ResponseWriter, r *http.Request) {
	f := models.AdminFilter{
		Jenis:    r.URL.Query().Get("jenis"),
		Status:   r.URL.Query().Get("status"),
		DateFrom: r.URL.Query().Get("date_from"),
		DateTo:   r.URL.Query().Get("date_to"),
		Search:   r.URL.Query().Get("q"),
	}

	requests, err := models.ListAdminEmailRequests(c.DB, f)
	if err != nil {
		http.Error(w, "Gagal memuat data, silakan coba lagi", http.StatusInternalServerError)
		return
	}
	totalRequests, err := models.CountEmailRequests(c.DB)
	if err != nil {
		http.Error(w, "Gagal menghitung usulan, silakan coba lagi", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Requests":      requests,
		"TotalRequests": totalRequests,
		"Filter":        f,
		"StatusOptions": models.StatusOptions,
		"JenisOptions":  models.JenisOptions,
		"CurrentQuery":  r.URL.RawQuery,
	}
	data["CSRFToken"], _ = middleware.CSRFToken(r)
	if err := c.Tmpl.ExecuteTemplate(w, "admin_dashboard.html", data); err != nil {
		http.Error(w, "Gagal menampilkan halaman", http.StatusInternalServerError)
	}
}

// UpdateStatus memproses form aksi "ubah status" dari dashboard.
func (c *AdminController) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !middleware.ValidCSRFToken(r) {
		http.Error(w, "Token keamanan tidak valid", http.StatusForbidden)
		return
	}

	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		http.Error(w, "ID tidak valid", http.StatusBadRequest)
		return
	}

	status := r.FormValue("status")
	valid := false
	for _, s := range models.StatusOptions {
		if s == status {
			valid = true
			break
		}
	}
	if !valid {
		http.Error(w, "Status tidak valid", http.StatusBadRequest)
		return
	}

	detail := r.FormValue("detail")
	if err := models.UpdateEmailRequestStatus(c.DB, id, status, detail); err != nil {
		http.Error(w, "Gagal memperbarui data", http.StatusInternalServerError)
		return
	}

	redirect := r.FormValue("redirect")
	if redirect == "" || !strings.HasPrefix(redirect, "/") || strings.HasPrefix(redirect, "//") {
		redirect = "/admin/dashboard"
	}
	http.Redirect(w, r, redirect, http.StatusSeeOther)
}

// DownloadFormulir mengunduh file formulir hanya untuk admin yang sudah login.
func (c *AdminController) DownloadFormulir(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, "ID tidak valid", http.StatusBadRequest)
		return
	}

	path, err := models.GetFormulirPath(c.DB, id)
	if err != nil {
		http.Error(w, "Formulir tidak ditemukan", http.StatusNotFound)
		return
	}
	cleanPath := filepath.Clean(path)
	if path == "" || filepath.Dir(cleanPath) != "uploads" || strings.Contains(cleanPath, "..") {
		http.Error(w, "Formulir tidak valid", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Disposition", `attachment; filename="formulir-pengajuan`+filepath.Ext(cleanPath)+`"`)
	http.ServeFile(w, r, cleanPath)
}
