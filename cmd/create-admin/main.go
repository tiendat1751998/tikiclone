package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

func sha256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

func main() {
	dsn := "tiki:tiki_dev@tcp(127.0.0.1:3306)/tiki_auth?charset=utf8mb4&parseTime=true&loc=UTC"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx := context.Background()
	email := "admin@admin.com"
	password := "Adm1n123*"
	displayName := "Root Admin"

	// Get admin role ID
	var roleID string
	err = db.QueryRowContext(ctx, "SELECT role_id FROM roles WHERE name = ?", "ADMIN").Scan(&roleID)
	if err != nil {
		log.Fatal("Find admin role failed:", err)
	}
	fmt.Printf("Admin role ID: %s\n", roleID)

	// Check if user exists in new users table
	var newUserID string
	err = db.QueryRowContext(ctx, "SELECT id FROM users WHERE email = ?", email).Scan(&newUserID)

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte(password), 12)
	now := time.Now().UTC()

	if err == nil && newUserID != "" {
		fmt.Printf("User exists in new users: %s\n", newUserID)
		// Update password and status
		_, _ = db.ExecContext(ctx, "UPDATE users SET password_hash = ?, status = 'active', email_verified = TRUE WHERE id = ?", string(passwordHash), newUserID)
	} else {
		// Create new user — use a deterministic ID so we can link to users_old
		newUserID = "usr-admin-001-root"
		_, err = db.ExecContext(ctx,
			`INSERT INTO users (id, email, email_hash, phone, username, password_hash, display_name, status, email_verified, phone_verified, mfa_enabled, twofa_secret, failed_attempts, metadata, created_at, updated_at)
			 VALUES (?, ?, ?, '', 'admin', ?, ?, 'active', TRUE, FALSE, FALSE, '', 0, '{}', ?, ?)`,
			newUserID, email, sha256Hex(email), string(passwordHash), displayName, now, now,
		)
		if err != nil {
			log.Fatal("Insert into users failed:", err)
		}
		fmt.Printf("Created in new users: %s\n", newUserID)
	}

	// Insert into users_old with SAME user_id
	var oldExists bool
	db.QueryRowContext(ctx, "SELECT COUNT(*) > 0 FROM users_old WHERE user_id = ?", newUserID).Scan(&oldExists)
	if !oldExists {
		_, err = db.ExecContext(ctx,
			`INSERT INTO users_old (user_id, email, phone, password_hash, full_name, role, is_verified, is_active, created_at, updated_at)
			 VALUES (?, ?, '', ?, ?, 'admin', 1, 1, ?, ?)`,
			newUserID, email, string(passwordHash), displayName, now, now,
		)
		if err != nil {
			fmt.Printf("Note: users_old insert: %v\n", err)
		} else {
			fmt.Printf("Inserted into users_old: %s\n", newUserID)
		}
	} else {
		fmt.Printf("Already exists in users_old: %s\n", newUserID)
	}

	// Assign admin role (FK references users_old.user_id)
	_, err = db.ExecContext(ctx,
		"INSERT IGNORE INTO user_roles (user_id, role_id) VALUES (?, ?)",
		newUserID, roleID,
	)
	if err != nil {
		fmt.Printf("Role assignment error: %v\n", err)
	} else {
		fmt.Println("Admin role assigned!")
	}

	// Verify
	var count int
	db.QueryRowContext(ctx, "SELECT COUNT(*) FROM user_roles WHERE user_id = ? AND role_id = ?", newUserID, roleID).Scan(&count)
	fmt.Printf("Role verification: %d rows\n", count)

	fmt.Printf("\n========================================\n")
	fmt.Printf("  ADMIN USER READY\n")
	fmt.Printf("========================================\n")
	fmt.Printf("  Email:    %s\n", email)
	fmt.Printf("  Password: Adm1n123*\n")
	fmt.Printf("  Login:    https://192.168.5.106:8443/login\n")
	fmt.Printf("========================================\n")
}
