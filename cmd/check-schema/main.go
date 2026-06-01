package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	dsn := "tiki:tiki_dev@tcp(127.0.0.1:3306)/tiki_auth?charset=utf8mb4&parseTime=true&loc=UTC"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx := context.Background()

	// Check users_old for admin
	fmt.Println("=== users_old admin entries ===")
	rows, _ := db.Query("SELECT user_id, email, role FROM users_old WHERE email = ?", "admin@admin.com")
	defer rows.Close()
	for rows.Next() {
		var uid, email, role string
		rows.Scan(&uid, &email, &role)
		fmt.Printf("  user_id=%s email=%s role=%s\n", uid, email, role)
	}

	// Check new users table
	fmt.Println("\n=== new users admin entries ===")
	rows2, _ := db.Query("SELECT id, email, status FROM users WHERE email = ?", "admin@admin.com")
	defer rows2.Close()
	for rows2.Next() {
		var id, email, status string
		rows2.Scan(&id, &email, &status)
		fmt.Printf("  id=%s email=%s status=%s\n", id, email, status)
	}

	// Check user_roles for all admin-related entries
	fmt.Println("\n=== user_roles for admin users ===")
	rows3, _ := db.QueryContext(ctx, `
		SELECT ur.user_id, ur.role_id, r.name 
		FROM user_roles ur 
		JOIN roles r ON r.role_id = ur.role_id 
		WHERE r.name = 'ADMIN'`)
	defer rows3.Close()
	for rows3.Next() {
		var uid, rid, name string
		rows3.Scan(&uid, &rid, &name)
		fmt.Printf("  user_id=%s role_id=%s role_name=%s\n", uid, rid, name)
	}

	// Get the admin role ID
	var roleID string
	db.QueryRowContext(ctx, "SELECT role_id FROM roles WHERE name = ?", "ADMIN").Scan(&roleID)
	fmt.Printf("\nAdmin role ID: %s\n", roleID)

	// Check which user_id in users_old has the admin email
	var oldUserID string
	db.QueryRowContext(ctx, "SELECT user_id FROM users_old WHERE email = ?", "admin@admin.com").Scan(&oldUserID)
	fmt.Printf("users_old admin user_id: %s\n", oldUserID)

	// Check if this user_id has the role
	var hasRole bool
	db.QueryRowContext(ctx, "SELECT COUNT(*) > 0 FROM user_roles WHERE user_id = ? AND role_id = ?", oldUserID, roleID).Scan(&hasRole)
	fmt.Printf("Has admin role: %v\n", hasRole)

	// If not, assign it
	if !hasRole {
		_, err = db.ExecContext(ctx, "INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)", oldUserID, roleID)
		if err != nil {
			fmt.Printf("Role assignment error: %v\n", err)
		} else {
			fmt.Println("Admin role assigned to users_old entry!")
		}
	}

	// Final verification
	fmt.Println("\n=== final verification ===")
	rows4, _ := db.QueryContext(ctx, `
		SELECT u.id, u.email, u.status, u.email_verified, ur.role_id, r.name
		FROM users u
		LEFT JOIN user_roles ur ON ur.user_id = u.id
		LEFT JOIN roles r ON r.role_id = ur.role_id
		WHERE u.email = ?`, "admin@admin.com")
	defer rows4.Close()
	for rows4.Next() {
		var id, email, status string
		var verified bool
		var roleID sql.NullString
		var roleName sql.NullString
		rows4.Scan(&id, &email, &status, &verified, &roleID, &roleName)
		fmt.Printf("  id=%s email=%s status=%s verified=%v role=%v role_name=%v\n", id, email, status, verified, roleID, roleName)
	}
}
