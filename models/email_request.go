package models

import (
	"database/sql"
	"time"
)

// EmailRequest merepresentasikan satu baris usulan pembuatan/perubahan akun
// email kampus, sekaligus jadi "Model" yang tahu cara mengambil & mengubah
// dirinya sendiri di database.
type EmailRequest struct {
	ID           int
	JenisUsulan  string // Dosen / Pegawai ITH, Mahasiswa ITH, Lembaga ITH
	Nama         string
	NipNim       string
	ProdiUnit    string
	FormulirPath string
	Status       string // Diajukan, Diproses, Selesai, Ditolak
	Detail       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// StatusOptions & JenisOptions dipakai untuk dropdown filter dan form aksi.
var StatusOptions = []string{"Diajukan", "Diproses", "Selesai", "Ditolak"}
var JenisOptions = []string{"Dosen / Pegawai ITH", "Mahasiswa ITH", "Lembaga / Unit ITH"}

// AdminFilter menampung parameter filter dashboard admin.
type AdminFilter struct {
	Jenis    string
	Status   string
	DateFrom string // format: YYYY-MM-DD
	DateTo   string
	Search   string
}

// nullable mengubah string kosong menjadi nil supaya aman dipakai pada
// perbandingan SQL seperti "$1::date" tanpa error cast.
func nullable(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func scanEmailRequests(rows *sql.Rows) ([]EmailRequest, error) {
	defer rows.Close()
	var out []EmailRequest
	for rows.Next() {
		var r EmailRequest
		var detail sql.NullString
		var formulirPath sql.NullString
		if err := rows.Scan(&r.ID, &r.JenisUsulan, &r.Nama, &r.NipNim, &r.ProdiUnit,
			&formulirPath, &r.Status, &detail, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		r.FormulirPath = formulirPath.String
		r.Detail = detail.String
		out = append(out, r)
	}
	return out, rows.Err()
}

// ListPublicEmailRequests mengambil maksimal 100 baris terbaru, dengan
// pencarian berdasarkan nama atau NIP/NIM. Tidak ada kolom sensitif yang
// dikembalikan (cukup untuk ditampilkan ke publik).
func ListPublicEmailRequests(db *sql.DB, search string, page, pageSize int) ([]EmailRequest, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize
	const q = `
		SELECT id, jenis_usulan, nama, nip_nim, prodi_unit, formulir_path, status, detail, created_at, updated_at
		FROM email_requests
		WHERE ($1 = '' OR nama ILIKE '%' || $1 || '%' OR nip_nim ILIKE '%' || $1 || '%')
		ORDER BY created_at ASC, id ASC
		LIMIT $2 OFFSET $3`
	rows, err := db.Query(q, search, pageSize, offset)
	if err != nil {
		return nil, err
	}
	return scanEmailRequests(rows)
}

// CountPublicEmailRequests menghitung usulan yang cocok dengan pencarian publik.
func CountPublicEmailRequests(db *sql.DB, search string) (int, error) {
	const q = `
		SELECT COUNT(*)
		FROM email_requests
		WHERE ($1 = '' OR nama ILIKE '%' || $1 || '%' OR nip_nim ILIKE '%' || $1 || '%')`
	var count int
	err := db.QueryRow(q, search).Scan(&count)
	return count, err
}

// ListAdminEmailRequests sama seperti versi public namun dengan filter
// tambahan: jenis usulan, status, dan rentang tanggal pengajuan.
func ListAdminEmailRequests(db *sql.DB, f AdminFilter, page, pageSize int) ([]EmailRequest, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize
	const q = `
		SELECT id, jenis_usulan, nama, nip_nim, prodi_unit, formulir_path, status, detail, created_at, updated_at
		FROM email_requests
		WHERE ($1::text IS NULL OR jenis_usulan = $1)
		  AND ($2::text IS NULL OR status = $2)
		  AND ($3::date IS NULL OR created_at >= $3::date)
		  AND ($4::date IS NULL OR created_at < ($4::date + INTERVAL '1 day'))
		  AND ($5 = '' OR nama ILIKE '%' || $5 || '%' OR nip_nim ILIKE '%' || $5 || '%')
		ORDER BY created_at ASC, id ASC
		LIMIT $6 OFFSET $7`
	rows, err := db.Query(q,
		nullable(f.Jenis), nullable(f.Status), nullable(f.DateFrom), nullable(f.DateTo), f.Search, pageSize, offset)
	if err != nil {
		return nil, err
	}
	return scanEmailRequests(rows)
}

// CountAdminEmailRequests menghitung data sesuai filter untuk pagination.
func CountAdminEmailRequests(db *sql.DB, f AdminFilter) (int, error) {
	const q = `
		SELECT COUNT(*)
		FROM email_requests
		WHERE ($1::text IS NULL OR jenis_usulan = $1)
		  AND ($2::text IS NULL OR status = $2)
		  AND ($3::date IS NULL OR created_at >= $3::date)
		  AND ($4::date IS NULL OR created_at < ($4::date + INTERVAL '1 day'))
		  AND ($5 = '' OR nama ILIKE '%' || $5 || '%' OR nip_nim ILIKE '%' || $5 || '%')`
	var count int
	err := db.QueryRow(q,
		nullable(f.Jenis), nullable(f.Status), nullable(f.DateFrom), nullable(f.DateTo), f.Search).Scan(&count)
	return count, err
}

// CountEmailRequests menghitung seluruh usulan yang sudah masuk.
func CountEmailRequests(db *sql.DB) (int, error) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM email_requests`).Scan(&count)
	return count, err
}

// UpdateEmailRequestStatus dipakai admin untuk mengubah status & detail
// satu usulan.
func UpdateEmailRequestStatus(db *sql.DB, id int, status, detail string) error {
	const q = `UPDATE email_requests SET status = $1, detail = $2, updated_at = NOW() WHERE id = $3`
	_, err := db.Exec(q, status, detail, id)
	return err
}

// CreateEmailRequest menyimpan usulan baru dari formulir publik pendaftaran.
// Status awal selalu "Diajukan" sampai ditindaklanjuti admin di dashboard.
func CreateEmailRequest(db *sql.DB, jenis, nama, nipNim, prodiUnit, formulirPath string) (int, error) {
	const q = `
		INSERT INTO email_requests (jenis_usulan, nama, nip_nim, prodi_unit, formulir_path, status)
		VALUES ($1, $2, $3, $4, $5, 'Diajukan')
		RETURNING id`
	var id int
	err := db.QueryRow(q, jenis, nama, nipNim, prodiUnit, formulirPath).Scan(&id)
	return id, err
}

// GetFormulirRequest mengambil metadata dan lokasi formulir untuk akses admin terproteksi.
func GetFormulirRequest(db *sql.DB, id int) (EmailRequest, error) {
	var request EmailRequest
	var path sql.NullString
	err := db.QueryRow(`
		SELECT jenis_usulan, nama, nip_nim, prodi_unit, formulir_path, created_at
		FROM email_requests
		WHERE id = $1`, id).Scan(
		&request.JenisUsulan, &request.Nama, &request.NipNim, &request.ProdiUnit,
		&path, &request.CreatedAt)
	request.FormulirPath = path.String
	return request, err
}

// IsValidJenis mengecek input jenis usulan dari formulir terhadap whitelist
// JenisOptions, supaya tidak ada nilai sembarangan yang masuk ke database.
func IsValidJenis(j string) bool {
	for _, v := range JenisOptions {
		if v == j {
			return true
		}
	}
	return false
}
