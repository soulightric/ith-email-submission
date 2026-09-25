package routes

import (
	"database/sql"
	"html/template"
	"net/http"

	"monitoring-email-ith/controllers"
	"monitoring-email-ith/middleware"
)

// Templates mengelompokkan semua template halaman yang sudah di-parse,
// supaya main.go tidak perlu tahu detail parsing-nya.
type Templates struct {
	Public       *template.Template
	Admin        *template.Template
	Login        *template.Template
	Registration *template.Template
}

// New merangkai seluruh Controller jadi satu http.Handler siap pakai.
// Ini satu-satunya tempat yang perlu dilihat untuk tahu endpoint apa saja
// yang tersedia dan mana yang dilindungi login.
func New(db *sql.DB, t Templates) http.Handler {
	public := controllers.NewPublicController(db, t.Public)
	admin := controllers.NewAdminController(db, t.Admin)
	auth := controllers.NewAuthController(db, t.Login)
	registration := controllers.NewRegistrationController(db, t.Registration)

	mux := http.NewServeMux()

	// --- Public, tanpa login ---
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("/", public.Index)
	mux.HandleFunc("/daftar", registration.Index)
	mux.HandleFunc("/login", auth.Login)
	mux.HandleFunc("/logout", auth.Logout)

	// --- Admin, wajib login (dibungkus middleware.RequireAuth) ---
	adminMux := http.NewServeMux()
	adminMux.HandleFunc("/admin/dashboard", admin.Dashboard)
	adminMux.HandleFunc("/admin/usulan/update", admin.UpdateStatus)
	adminMux.HandleFunc("/admin/usulan/formulir", admin.DownloadFormulir)
	mux.Handle("/admin/", middleware.RequireAuth(adminMux))

	return securityHeaders(mux)
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		next.ServeHTTP(w, r)
	})
}
