package models

import "database/sql"

type User struct {
	ID           int
	Username     string
	PasswordHash string
	Role         string
}

// FindUserByUsername dipakai proses login.
func FindUserByUsername(db *sql.DB, username string) (User, error) {
	var u User
	err := db.QueryRow(
		`SELECT id, username, password_hash, role FROM users WHERE username = $1`, username,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role)
	return u, err
}
