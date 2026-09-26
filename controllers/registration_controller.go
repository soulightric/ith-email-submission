package controllers

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"monitoring-email-ith/models"
)

// RegistrationController menangani formulir pengajuan usulan akun email
// yang bisa diisi siapa saja tanpa login, mirip alur Google Form: submit
// lalu tampil halaman konfirmasi di URL yang sama.
type RegistrationController struct {
	DB   *sql.DB
	Tmpl *template.Template
}

func NewRegistrationController(db *sql.DB, tmpl *template.Template) *RegistrationController {
	return &RegistrationController{DB: db, Tmpl: tmpl}
}

type registrationForm struct {
	JenisOptions []string
	Jenis        string
	Nama         string
	NipNim       string
	ProdiUnit    string
	FormulirName string
	Error        string
	Success      bool
}

// Index menampilkan formulir (GET) dan memproses pengiriman (POST).
func (c *RegistrationController) Index(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		c.render(w, registrationForm{JenisOptions: models.JenisOptions})
		return
	}

	data := registrationForm{
		JenisOptions: models.JenisOptions,
		Jenis:        r.FormValue("jenis_usulan"),
		Nama:         strings.TrimSpace(r.FormValue("nama")),
		NipNim:       strings.TrimSpace(r.FormValue("nip_nim")),
		ProdiUnit:    strings.TrimSpace(r.FormValue("prodi_unit")),
	}

	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		data.Error = "Ukuran formulir terlalu besar. Maksimal 10 MB."
		c.render(w, data)
		return
	}

	file, header, err := r.FormFile("formulir")
	if err != nil {
		data.Error = "Formulir pengajuan wajib diunggah."
		c.render(w, data)
		return
	}
	defer file.Close()
	data.FormulirName = header.Filename
	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowedExtensions := map[string]bool{".pdf": true, ".doc": true, ".docx": true}
	if !allowedExtensions[ext] {
		data.Error = "Format formulir harus PDF, DOC, atau DOCX."
		c.render(w, data)
		return
	}
	contentHeader := make([]byte, 512)
	contentSize, readErr := file.Read(contentHeader)
	if readErr != nil && readErr != io.EOF {
		data.Error = "Gagal membaca formulir."
		c.render(w, data)
		return
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		data.Error = "Gagal memproses formulir."
		c.render(w, data)
		return
	}
	contentType := http.DetectContentType(contentHeader[:contentSize])
	validContentType := (ext == ".pdf" && contentType == "application/pdf") ||
		(ext == ".doc" && contentType == "application/msword") ||
		(ext == ".docx" && (contentType == "application/zip" || bytes.HasPrefix(contentHeader[:contentSize], []byte("PK"))))
	if !validContentType {
		data.Error = "Isi file tidak sesuai dengan format yang dipilih."
		c.render(w, data)
		return
	}

	switch {
	case !models.IsValidJenis(data.Jenis):
		data.Error = "Pilih jenis usulan yang valid."
	case data.Nama == "" || data.NipNim == "" || data.ProdiUnit == "":
		data.Error = "Semua kolom wajib diisi."
	}
	if data.Error != "" {
		c.render(w, data)
		return
	}

	if err := os.MkdirAll("uploads", 0750); err != nil {
		data.Error = "Gagal menyiapkan penyimpanan formulir."
		c.render(w, data)
		return
	}

	randomName := make([]byte, 16)
	if _, err := rand.Read(randomName); err != nil {
		data.Error = "Gagal menyiapkan nama file formulir."
		c.render(w, data)
		return
	}
	storedBasePath := filepath.Join("uploads", hex.EncodeToString(randomName))
	storedPath, err := storeUploadedDocument(file, storedBasePath, ext, header.Size)
	if err != nil {
		data.Error = "Gagal menyimpan formulir."
		c.render(w, data)
		return
	}

	if _, err := models.CreateEmailRequest(c.DB, data.Jenis, data.Nama, data.NipNim, data.ProdiUnit, storedPath); err != nil {
		_ = os.Remove(storedPath)
		data.Error = "Gagal menyimpan usulan, silakan coba lagi."
		c.render(w, data)
		return
	}

	c.render(w, registrationForm{JenisOptions: models.JenisOptions, Success: true})
}

func (c *RegistrationController) render(w http.ResponseWriter, data registrationForm) {
	if err := c.Tmpl.ExecuteTemplate(w, "daftar.html", data); err != nil {
		http.Error(w, "Gagal menampilkan halaman", http.StatusInternalServerError)
	}
}
