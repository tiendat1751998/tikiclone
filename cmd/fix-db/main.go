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

	// Check the admin user in new users table
	var id, email, username, passwordHash, status string
	var emailVerified bool
	err = db.QueryRowContext(ctx,
		"SELECT id, email, username, password_hash, status, email_verified FROM users WHERE email = ?",
		"admin@admin.com",
	).Scan(&id, &email, &username, &passwordHash, &status, &emailVerified)
	if err != nil {
		log.Fatal("Query failed:", err)
	}

	fmt.Printf("User: id=%s\n", id)
	fmt.Printf("  email=%s\n", email)
	fmt.Printf("  username=%s\n", username)
	fmt.Printf("  status=%s\n", status)
	fmt.Printf("  email_verified=%v\n", emailVerified)
	fmt.Printf("  password_hash prefix=%s\n", passwordHash[:30])
	fmt.Printf("  hash length=%d\n", len(passwordHash))

	// Check if hash is valid bcrypt
	if len(passwordHash) >= 4 && (passwordHash[:4] == "$2a$" || passwordHash[:4] == "$2b$") {
		fmt.Println("  hash type: bcrypt ✓")
		// Verify the password
		err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte("Adm1n123*"))
		if err != nil {
			fmt.Printf("  bcrypt verify: FAILED - %v\n", err)
		} else {
			fmt.Println("  bcrypt verify: OK ✓")
		}
	} else {
		fmt.Println("  hash type: NOT bcrypt ✗")
	}

	// Check CanLogin conditions
	fmt.Printf("\nCanLogin checks:\n")
	fmt.Printf("  status == 'active'? %v\n", status == "active")
	fmt.Printf("  email_verified? %v\n", emailVerified)

	// Check if there's a locked_until issue
	var lockedUntil sql.NullTime
	var failedAttempts int
	db.QueryRowContext(ctx, "SELECT locked_until, failed_attempts FROM users WHERE id = ?", id).Scan(&lockedUntil, &failedAttempts)
	fmt.Printf("  locked_until: %v\n", lockedUntil)
	fmt.Printf("  failed_attempts: %d\n", failedAttempts)

	// Check sessions table
	var sessionCount int
	db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sessions WHERE user_id = ?", id).Scan(&sessionCount)
	fmt.Printf("  existing sessions: %d\n", sessionCount)

	// Now let's try to fix any issues
	// 1. Make sure status is active
	// 2. Make sure email_verified is true
	// 3. Reset failed_attempts
	// 4. Clear locked_until
	// 5. Generate a fresh bcrypt hash

	fmt.Println("\n=== FIXING USER ===")
	freshHash, _ := bcrypt.GenerateFromPassword([]byte("Adm1n123*"), 12)
	now := time.Now().UTC()

	_, err = db.ExecContext(ctx,
		`UPDATE users SET 
			password_hash = ?, 
			status = 'active', 
			email_verified = TRUE, 
			failed_attempts = 0, 
			locked_until = NULL,
			updated_at = ?
		WHERE id = ?`,
		string(freshHash), now, id,
	)
	if err != nil {
		log.Fatal("Update failed:", err)
	}
	fmt.Println("User fixed!")

	// Verify the fix
	var newHash string
	db.QueryRowContext(ctx, "SELECT password_hash FROM users WHERE id = ?", id).Scan(&newHash)
	err = bcrypt.CompareHashAndPassword([]byte(newHash), []byte("Adm1n123*"))
	if err != nil {
		fmt.Printf("Fresh hash verify: FAILED - %v\n", err)
	} else {
		fmt.Println("Fresh hash verify: OK ✓")
	}

	// Also check the users_old table
	fmt.Println("\n=== users_old check ===")
	var oldID, oldEmail, oldStatus string
	err = db.QueryRowContext(ctx, "SELECT user_id, email, role FROM users_old WHERE email = ?", "admin@admin.com").Scan(&oldID, &oldEmail, &oldStatus)
	if err != nil {
		fmt.Printf("users_old: %v\n", err)
	} else {
		fmt.Printf("  user_id=%s email=%s role=%s\n", oldID, oldEmail, oldStatus)
	}

	// Check user_roles
	fmt.Println("\n=== user_roles check ===")
	rows, _ := db.QueryContext(ctx, "SELECT user_id, role_id FROM user_roles WHERE user_id = ?", id)
	defer rows.Close()
	found := false
	for rows.Next() {
		var uid, rid string
		rows.Scan(&uid, &rid)
		fmt.Printf("  user_id=%s role_id=%s\n", uid, rid)
		found = true
	}
	if !found {
		fmt.Println("  NO ROLES FOUND!")
	}

	fmt.Println("\n=== DONE ===")
	fmt.Println("Try logging in again now.")
}
