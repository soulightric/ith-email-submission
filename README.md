# Monitoring Pembuatan Email Kampus ITH (struktur MVC)

Aplikasi Go (net/http + html/template) untuk memantau status usulan
pembuatan/perubahan akun email kampus, disusun dengan pola MVC supaya
tanggung jawab tiap bagian jelas.

## Struktur

```
main.go              -> wiring saja: load config, konek DB, parse view, jalankan server
config/               -> baca environment variable
database/             -> koneksi PostgreSQL (connection pool)
models/               -> [MODEL] struct data + semua query SQL terkait tabelnya
  email_request.go
  user.go
controllers/           -> [CONTROLLER] terima request, panggil Model, render View
  public_controller.go
  admin_controller.go
  auth_controller.go
middleware/            -> session store & proteksi route admin (dipakai lintas controller)
routes/                -> [ROUTES] satu tempat untuk lihat semua endpoint & mana yang diproteksi
views/                 -> [VIEW] file HTML (html/template, auto-escape aktif)
static/                -> CSS + JS (modal detail & modal ubah status)
schema.sql             -> skema tabel PostgreSQL
```

Alur satu request, misalnya buka dashboard admin:

```
routes.go -> middleware.RequireAuth -> AdminController.Dashboard
           -> models.ListAdminEmailRequests (query ke DB)
           -> render views/admin_dashboard.html
```

Kalau nanti mau nambah fitur, aturan sederhananya:
- Nambah tabel/kolom baru / query baru → kerja di `models/`
- Nambah endpoint baru → kerja di `controllers/` + daftarkan di `routes/routes.go`
- Nambah halaman baru → kerja di `views/`
- Ubah aturan login/akses → kerja di `middleware/`

## Setup

1. Buat database dan jalankan skema:
   ```bash
  createdb monitoring_email
  psql -d monitoring_email -f schema.sql
   ```

2. Buat akun admin. Generate hash bcrypt dulu, misalnya lewat skrip kecil:
   ```bash
   cat <<'EOF' > /tmp/hash.go
   package main
   import (
       "fmt"
       "golang.org/x/crypto/bcrypt"
   )
   func main() {
       h, _ := bcrypt.GenerateFromPassword([]byte("password-anda"), bcrypt.DefaultCost)
       fmt.Println(string(h))
   }
   EOF
   go run /tmp/hash.go
   ```
   Lalu masukkan ke tabel `users`:
   ```sql
   INSERT INTO users (username, password_hash, role)
   VALUES ('admin', '<hash-hasil-generate>', 'admin');
   ```

3. Set environment variable (opsional, ada default):
   ```
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=postgres
   DB_PASSWORD=postgres
   DB_NAME=monitoring_email
   DB_SSLMODE=disable
   PORT=8080
   ```

4. Ambil dependency dan jalankan:
   ```bash
   go mod tidy
   go run .
   ```

  Saat aplikasi berjalan, kolom upload formulir untuk database lama akan
  disiapkan otomatis. File upload disimpan di folder `uploads/` dan tidak
  dimasukkan ke repository.

5. Buka:
   - Halaman publik: `http://localhost:8080/`
   - Login admin: `http://localhost:8080/login`
   - Dashboard admin: `http://localhost:8080/admin/dashboard`

## Catatan keamanan

- Cookie session diberi flag `HttpOnly`, `Secure`, dan `SameSite=Strict`.
  Saat development tanpa HTTPS, set `Secure: false` sementara di
  `controllers/auth_controller.go`, atau jalankan di belakang reverse
  proxy TLS (mis. Nginx).
- Password disimpan sebagai hash `bcrypt`, tidak pernah plaintext.
- Semua route `/admin/*` dibungkus `middleware.RequireAuth` — dicek di
  `routes/routes.go`, bukan tersebar di tiap handler.
- Semua query memakai parameterized query (`$1`, `$2`, ...) di `models/`,
  bukan string concatenation, untuk mencegah SQL injection.
- `html/template` (bukan `text/template`) dipakai di semua view sehingga
  output otomatis di-escape, mencegah XSS lewat data seperti nama atau
  catatan detail.
- Session saat ini disimpan in-memory (`middleware/auth.go`) — cukup
  untuk single instance. Kalau nanti deploy multi instance atau butuh
  session tahan restart, pindahkan ke tabel `sessions` di PostgreSQL
  atau Redis.
- Belum ada rate limiting di endpoint login — tambahkan kalau aplikasi
  ini publik dan rawan brute force.

## Fitur

**Halaman publik (`/`)**
- Tabel: No, Jenis Usulan, Nama, NIP/NIM, Prodi/Unit, Status
- Klik status untuk melihat detail (modal)
- Pencarian berdasarkan nama atau NIP/NIM
- Dibatasi 100 baris terbaru
- Tombol "+ Ajukan Usulan Baru" menuju formulir pendaftaran

**Formulir pendaftaran (`/daftar`)**
- Bergaya Google Form: header berwarna, satu kolom, konfirmasi setelah submit
- Jenis email: Dosen / Pegawai ITH, Mahasiswa ITH, atau Lembaga ITH
- Field: Jenis Email, Nama, NIP/NIM, Program Studi/Unit, dan upload formulir pengajuan
- Upload menerima PDF, DOC, atau DOCX dengan ukuran maksimal 10 MB
- Untuk Mahasiswa ITH, tampil rekomendasi nama email `nama.nim` tanpa `@ith.ac.id`
- Tidak perlu login; status awal usulan otomatis "Diajukan"
- Validasi server-side: semua field wajib diisi, jenis usulan divalidasi terhadap whitelist

**Dashboard admin (`/admin/dashboard`, perlu login)**
- Tabel sama seperti publik + kolom Aksi (ubah status & catatan)
- Filter: jenis usulan, status, rentang tanggal pengajuan, pencarian nama/NIM
- Update status lewat modal, redirect kembali dengan filter yang sama
