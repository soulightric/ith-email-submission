package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"sync"
	"time"
)

// CookieName adalah nama cookie session yang dipakai di seluruh aplikasi.
const CookieName = "session_token"

// TTL adalah masa berlaku session sejak login.
const TTL = 8 * time.Hour

type session struct {
	UserID    int
	Username  string
	CSRFToken string
	Expires   time.Time
}

// Session store in-memory. Untuk deployment multi-instance, ganti dengan
// tabel "sessions" di PostgreSQL atau Redis supaya session tidak hilang
// saat restart / tidak sticky ke satu instance saja.
var store = struct {
	sync.Mutex
	m map[string]session
}{m: make(map[string]session)}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// CreateSession membuat token session baru untuk user yang berhasil login.
func CreateSession(userID int, username string) (token string, expires time.Time, err error) {
	token, err = generateToken()
	if err != nil {
		return "", time.Time{}, err
	}
	expires = time.Now().Add(TTL)
	csrfToken, err := generateToken()
	if err != nil {
		return "", time.Time{}, err
	}
	store.Lock()
	store.m[token] = session{UserID: userID, Username: username, CSRFToken: csrfToken, Expires: expires}
	store.Unlock()
	return token, expires, nil
}

// CSRFToken mengambil token anti-CSRF dari session yang sedang aktif.
func CSRFToken(r *http.Request) (string, bool) {
	cookie, err := r.Cookie(CookieName)
	if err != nil {
		return "", false
	}
	store.Lock()
	sess, ok := store.m[cookie.Value]
	store.Unlock()
	if !ok || time.Now().After(sess.Expires) {
		return "", false
	}
	return sess.CSRFToken, true
}

// ValidCSRFToken membandingkan token request dengan token session secara konstan.
func ValidCSRFToken(r *http.Request) bool {
	expected, ok := CSRFToken(r)
	provided := r.FormValue("csrf_token")
	if !ok || provided == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(expected), []byte(provided)) == 1
}

// DestroySession menghapus session saat logout.
func DestroySession(token string) {
	store.Lock()
	delete(store.m, token)
	store.Unlock()
}

// RequireAuth melindungi semua route di bawah /admin/. Mengecek cookie
// session valid dan belum kedaluwarsa sebelum meneruskan ke handler asli.
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(CookieName)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		store.Lock()
		sess, ok := store.m[c.Value]
		store.Unlock()
		if !ok || time.Now().After(sess.Expires) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}
