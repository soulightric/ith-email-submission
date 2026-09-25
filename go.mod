module monitoring-email-ith

go 1.22

require (
	github.com/joho/godotenv v1.5.1
	github.com/lib/pq v1.10.9
	golang.org/x/crypto v0.27.0
)

// Mirror GitHub dipakai sebagai fallback kalau golang.org tidak bisa
// diakses langsung di jaringan tertentu (proxy/firewall korporat, CI, dsb).
// Baris ini aman dihapus di lingkungan yang bisa akses golang.org normal.
replace golang.org/x/crypto => github.com/golang/crypto v0.27.0
