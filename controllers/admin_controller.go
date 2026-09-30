package controllers

import (
	"compress/gzip"
	"database/sql"
	"html/template"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"

	"monitoring-email-ith/mailer"
	"monitoring-email-ith/middleware"
	"monitoring-email-ith/models"
)

// AdminController menangani dashboard yang butuh login (dilindungi
// middleware.RequireAuth di routes).
type AdminController struct {
	DB   *sql.DB
	Tmpl *template.Template
	Mail mailer.Config
}

func NewAdminController(db *sql.DB, tmpl *template.Template, mailConfig mailer.Config) *AdminController {
	return &AdminController{DB: db, Tmpl: tmpl, Mail: mailConfig}
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
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = 1
	}
	const pageSize = 100

	requests, err := models.ListAdminEmailRequests(c.DB, f, page, pageSize)
	if err != nil {
		http.Error(w, "Gagal memuat data, silakan coba lagi", http.StatusInternalServerError)
		return
	}
	filteredTotal, err := models.CountAdminEmailRequests(c.DB, f)
	if err != nil {
		http.Error(w, "Gagal menghitung halaman, silakan coba lagi", http.StatusInternalServerError)
		return
	}
	totalRequests, err := models.CountEmailRequests(c.DB)
	if err != nil {
		http.Error(w, "Gagal menghitung usulan, silakan coba lagi", http.StatusInternalServerError)
		return
	}

	totalPages := (filteredTotal + pageSize - 1) / pageSize
	if totalPages > 0 && page > totalPages {
		page = totalPages
		requests, err = models.ListAdminEmailRequests(c.DB, f, page, pageSize)
		if err != nil {
			http.Error(w, "Gagal memuat halaman, silakan coba lagi", http.StatusInternalServerError)
			return
		}
	}
	pageURL := func(number int) string {
		values := url.Values{}
		if f.Search != "" {
			values.Set("q", f.Search)
		}
		if f.Jenis != "" {
			values.Set("jenis", f.Jenis)
		}
		if f.Status != "" {
			values.Set("status", f.Status)
		}
		if f.DateFrom != "" {
			values.Set("date_from", f.DateFrom)
		}
		if f.DateTo != "" {
			values.Set("date_to", f.DateTo)
		}
		values.Set("page", strconv.Itoa(number))
		return "/admin/dashboard?" + values.Encode()
	}
	pageLinks := make([]map[string]interface{}, 0, totalPages)
	for number := 1; number <= totalPages; number++ {
		pageLinks = append(pageLinks, map[string]interface{}{"Number": number, "URL": pageURL(number), "Current": number == page})
	}

	data := map[string]interface{}{
		"Requests":      requests,
		"TotalRequests": totalRequests,
		"FilteredTotal": filteredTotal,
		"Page":          page,
		"PageOffset":    (page - 1) * pageSize,
		"TotalPages":    totalPages,
		"HasPrevious":   page > 1,
		"PreviousPage":  page - 1,
		"PreviousURL":   pageURL(page - 1),
		"HasNext":       page < totalPages,
		"NextPage":      page + 1,
		"NextURL":       pageURL(page + 1),
		"PageLinks":     pageLinks,
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
	if status == "Selesai" {
		request, err := models.GetEmailRequestContact(c.DB, id)
		if err != nil {
			http.Error(w, "Pengajuan tidak ditemukan", http.StatusNotFound)
			return
		}
		if request.Status != "Selesai" {
			accountEmail := strings.TrimSpace(r.FormValue("account_email"))
			password := r.FormValue("login_password")
			destination := strings.TrimSpace(r.FormValue("contact_email"))
			if request.ContactEmail != "" {
				destination = request.ContactEmail
			}
			if !validRequestAccountEmail(accountEmail, request.JenisUsulan) || !validEmailAddress(destination) || password == "" || strings.ContainsAny(password, "\r\n") {
				http.Error(w, "Domain email akun harus sesuai jenis pengajuan; email penerima dan password yang valid wajib diisi saat menyelesaikan pengajuan", http.StatusBadRequest)
				return
			}
			if err := mailer.SendCredentials(c.Mail, destination, accountEmail, password); err != nil {
				log.Printf("Pengiriman kredensial SMTP gagal untuk pengajuan %d: %v", id, err)
				http.Error(w, "Email kredensial gagal dikirim. Periksa konfigurasi SMTP lalu coba lagi.", http.StatusBadGateway)
				return
			}
		}
	}
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

func validRequestAccountEmail(value, requestType string) bool {
	if !validEmailAddress(value) {
		return false
	}
	if requestType == "Mahasiswa ITH" {
		return strings.HasSuffix(strings.ToLower(value), "@mahasiswa.ith.ac.id")
	}
	return strings.HasSuffix(strings.ToLower(value), "@ith.ac.id")
}

// DownloadFormulir mengunduh file formulir hanya untuk admin yang sudah login.
func (c *AdminController) DownloadFormulir(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, "ID tidak valid", http.StatusBadRequest)
		return
	}

	request, err := models.GetFormulirRequest(c.DB, id)
	if err != nil {
		http.Error(w, "Formulir tidak ditemukan", http.StatusNotFound)
		return
	}
	cleanPath := filepath.Clean(request.FormulirPath)
	if request.FormulirPath == "" || filepath.Dir(cleanPath) != "uploads" || strings.Contains(cleanPath, "..") {
		http.Error(w, "Formulir tidak valid", http.StatusNotFound)
		return
	}

	documentExtension, compressed := storedDocumentExtension(cleanPath)
	filename := formulirDownloadFilename(request, documentExtension)
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": filename}))
	if compressed {
		storedFile, err := os.Open(cleanPath)
		if err != nil {
			http.Error(w, "Formulir tidak ditemukan", http.StatusNotFound)
			return
		}
		defer storedFile.Close()

		reader, err := gzip.NewReader(storedFile)
		if err != nil {
			http.Error(w, "Formulir tidak dapat dibaca", http.StatusInternalServerError)
			return
		}
		defer reader.Close()

		contentType := mime.TypeByExtension(documentExtension)
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		w.Header().Set("Content-Type", contentType)
		_, _ = io.Copy(w, reader)
		return
	}
	http.ServeFile(w, r, cleanPath)
}

func storedDocumentExtension(path string) (string, bool) {
	storedExtension := filepath.Ext(path)
	if strings.EqualFold(storedExtension, ".gz") {
		originalPath := path[:len(path)-len(storedExtension)]
		return filepath.Ext(originalPath), true
	}
	return storedExtension, false
}

func formulirDownloadFilename(request models.EmailRequest, extension string) string {
	parts := []string{
		"Formulir-Pengajuan",
		safeFilenamePart(request.Nama),
		safeFilenamePart(request.JenisUsulan),
		safeFilenamePart(request.NipNim),
		safeFilenamePart(request.ProdiUnit),
		request.CreatedAt.Format("2006-01-02"),
	}
	return strings.Join(parts, "_") + strings.ToLower(extension)
}

func safeFilenamePart(value string) string {
	normalized := strings.Map(func(char rune) rune {
		switch char {
		case '/', '\\', ':', '*', '?', '"', '<', '>', '|':
			return ' '
		default:
			if unicode.IsControl(char) {
				return -1
			}
			return char
		}
	}, value)
	part := strings.Join(strings.Fields(normalized), "-")
	part = strings.Trim(part, ".-")
	if part == "" {
		return "tanpa-data"
	}
	return part
}
