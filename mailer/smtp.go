package mailer

import (
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"strconv"
	"strings"
)

type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

func SendCredentials(config Config, recipient, accountEmail, password string) error {
	if config.Host == "" || config.Port < 1 || config.Port > 65535 || config.Username == "" || config.Password == "" {
		return fmt.Errorf("konfigurasi SMTP host, port, dan autentikasi wajib diisi")
	}
	if !validAddress(recipient) || !validAddress(accountEmail) || !validAddress(config.From) {
		return fmt.Errorf("alamat email tidak valid")
	}
	lowerAccountEmail := strings.ToLower(accountEmail)
	if (!strings.HasSuffix(lowerAccountEmail, "@ith.ac.id") && !strings.HasSuffix(lowerAccountEmail, "@mahasiswa.ith.ac.id")) || password == "" || strings.ContainsAny(password, "\r\n") {
		return fmt.Errorf("kredensial akun tidak valid")
	}

	auth := smtp.PlainAuth("", config.Username, config.Password, config.Host)
	message := strings.Join([]string{
		"To: " + recipient,
		"From: UPT TIK No Reply <" + config.From + ">",
		"Subject: Kredensial akun email ITH",
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		"Yth. pemohon,",
		"",
		"Pengajuan akun email ITH Anda telah selesai dan disetujui oleh UPT TIK ITH, berikut adalah nama email dan kredensialnya.",
		"",
		"Nama email: " + accountEmail,
		"Password login: " + password,
		"",
		"Pesan ini dikirim otomatis.",
	}, "\r\n")
	address := net.JoinHostPort(config.Host, strconv.Itoa(config.Port))
	return smtp.SendMail(address, auth, config.From, []string{recipient}, []byte(message))
}

func validAddress(value string) bool {
	address, err := mail.ParseAddress(value)
	return err == nil && address.Address == value
}
