package main

import (
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/joho/godotenv"

	"monitoring-email-ith/config"
	"monitoring-email-ith/database"
	"monitoring-email-ith/mailer"
	"monitoring-email-ith/routes"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("file .env tidak ditemukan, memakai environment variable sistem: %v", err)
	}

	cfg := config.Load()

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("gagal konek database: %v", err)
	}
	defer db.Close()

	funcMap := template.FuncMap{
		"lower": strings.ToLower,
		"inc":   func(i int) int { return i + 1 },
		"add":   func(a, b int) int { return a + b },
	}

	tmpl := routes.Templates{
		Public:       template.Must(template.New("public_index.html").Funcs(funcMap).ParseFiles("views/public_index.html")),
		Admin:        template.Must(template.New("admin_dashboard.html").Funcs(funcMap).ParseFiles("views/admin_dashboard.html")),
		Login:        template.Must(template.New("login.html").Funcs(funcMap).ParseFiles("views/login.html")),
		Registration: template.Must(template.New("daftar.html").Funcs(funcMap).ParseFiles("views/daftar.html")),
	}

	handler := routes.New(db, tmpl, mailer.Config{
		Host: cfg.SMTPHost, Port: cfg.SMTPPort, Username: cfg.SMTPUser,
		Password: cfg.SMTPPass, From: cfg.SMTPFrom,
	})

	addr := ":" + cfg.Port
	log.Printf("Server berjalan di http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, handler))
}
