package main

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"image/png"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	log "github.com/sirupsen/logrus"
	_ "github.com/mattn/go-sqlite3"
)

// TOTPSecret represents a user's TOTP secret
type TOTPSecret struct {
	Username string
	Secret   string
	Enabled  bool
}

// generateTOTPSecret generates a new TOTP secret for a user
func generateTOTPSecret(username string) (*otp.Key, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "ovpn-admin",
		AccountName: username,
	})
	if err != nil {
		return nil, err
	}
	return key, nil
}

// generateTOTPQRCode generates a QR code image for TOTP enrollment
func generateTOTPQRCode(key *otp.Key) (string, error) {
	var buf bytes.Buffer
	img, err := key.Image(200, 200)
	if err != nil {
		return "", err
	}
	
	err = png.Encode(&buf, img)
	if err != nil {
		return "", err
	}
	
	// Return base64 encoded PNG
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

// verifyTOTP verifies a TOTP code for a given secret
func verifyTOTP(secret, code string) bool {
	return totp.Validate(code, secret)
}

// initTOTPDB initializes the TOTP database table in the users database
func initTOTPDB(dbPath string) error {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	createTableSQL := `CREATE TABLE IF NOT EXISTS totp_secrets (
		username TEXT PRIMARY KEY,
		secret TEXT NOT NULL,
		enabled INTEGER NOT NULL DEFAULT 0
	);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		return err
	}

	log.Debug("TOTP table initialized in users database")
	return nil
}

// saveTOTPSecret saves a TOTP secret for a user
func saveTOTPSecret(dbPath, username, secret string, enabled bool) error {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	enabledInt := 0
	if enabled {
		enabledInt = 1
	}

	stmt, err := db.Prepare("INSERT OR REPLACE INTO totp_secrets(username, secret, enabled) VALUES(?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(username, secret, enabledInt)
	return err
}

// getTOTPSecret retrieves the TOTP secret for a user
func getTOTPSecret(dbPath, username string) (*TOTPSecret, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	row := db.QueryRow("SELECT username, secret, enabled FROM totp_secrets WHERE username = ?", username)
	
	var totpSecret TOTPSecret
	var enabledInt int
	err = row.Scan(&totpSecret.Username, &totpSecret.Secret, &enabledInt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No TOTP secret found
		}
		return nil, err
	}

	totpSecret.Enabled = enabledInt == 1
	return &totpSecret, nil
}

// deleteTOTPSecret deletes the TOTP secret for a user
func deleteTOTPSecret(dbPath, username string) error {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	stmt, err := db.Prepare("DELETE FROM totp_secrets WHERE username = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(username)
	return err
}

// enableTOTP enables TOTP for a user
func enableTOTP(dbPath, username string) error {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	stmt, err := db.Prepare("UPDATE totp_secrets SET enabled = 1 WHERE username = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(username)
	return err
}

// disableTOTP disables TOTP for a user
func disableTOTP(dbPath, username string) error {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	stmt, err := db.Prepare("UPDATE totp_secrets SET enabled = 0 WHERE username = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(username)
	return err
}
