package controllers

import (
	"database/sql"
	"html/template"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"

	"monitoring-email-ith/middleware"
	"monitoring-email-ith/models"
)

// AuthController menangani login dan logout admin.
type AuthController struct {
	DB   *sql.DB
	Tmpl *template.Template
}

const maxLoginFailures = 5
const loginFailureWindow = 15 * time.Minute

var loginFailures = struct {
	sync.Mutex
	m map[string][]time.Time
}{m: make(map[string][]time.Time)}

func NewAuthController(db *sql.DB, tmpl *template.Template) *AuthController {
	return &AuthController{DB: db, Tmpl: tmpl}
}

// Login menangani GET (tampilkan form) dan POST (proses login).
func (c *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		_ = c.Tmpl.ExecuteTemplate(w, "login.html", nil)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)

	username := r.FormValue("username")
	password := r.FormValue("password")
	clientKey := loginClientKey(r, username)
	if loginBlocked(clientKey) {
		w.WriteHeader(http.StatusTooManyRequests)
		_ = c.Tmpl.ExecuteTemplate(w, "login.html", map[string]string{"Error": "Terlalu banyak percobaan. Coba lagi nanti."})
		return
	}

	user, err := models.FindUserByUsername(c.DB, username)
	if err != nil || user.Role != "admin" || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		recordLoginFailure(clientKey)
		w.WriteHeader(http.StatusUnauthorized)
		_ = c.Tmpl.ExecuteTemplate(w, "login.html", map[string]string{
			"Error": "Username atau password salah",
		})
		return
	}
	clearLoginFailures(clientKey)

	token, expires, err := middleware.CreateSession(user.ID, user.Username)
	if err != nil {
		http.Error(w, "Terjadi kesalahan, coba lagi", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     middleware.CookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true, // set false hanya saat development tanpa HTTPS
		SameSite: http.SameSiteStrictMode,
		Expires:  expires,
	})
	http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
}

func loginClientKey(r *http.Request, username string) string {
	clientIP := r.RemoteAddr
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		clientIP = host
	}
	return clientIP + "\x00" + strings.ToLower(strings.TrimSpace(username))
}

func loginBlocked(key string) bool {
	now := time.Now()
	loginFailures.Lock()
	defer loginFailures.Unlock()
	failures := loginFailures.m[key]
	kept := failures[:0]
	for _, failure := range failures {
		if now.Sub(failure) < loginFailureWindow {
			kept = append(kept, failure)
		}
	}
	loginFailures.m[key] = kept
	return len(kept) >= maxLoginFailures
}

func recordLoginFailure(key string) {
	loginFailures.Lock()
	loginFailures.m[key] = append(loginFailures.m[key], time.Now())
	loginFailures.Unlock()
}

func clearLoginFailures(key string) {
	loginFailures.Lock()
	delete(loginFailures.m, key)
	loginFailures.Unlock()
}

// Logout menghapus session dan cookie.
func (c *AuthController) Logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(middleware.CookieName); err == nil {
		middleware.DestroySession(cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: middleware.CookieName, Value: "", Path: "/", MaxAge: -1})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
