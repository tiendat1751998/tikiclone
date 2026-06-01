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

	// Get admin user ID
	var userID string
	err = db.QueryRowContext(ctx, "SELECT id FROM users WHERE email = ?", "admin@admin.com").Scan(&userID)
	if err != nil {
		log.Fatal("Find admin failed:", err)
	}
	fmt.Printf("Admin user ID: %s\n", userID)

	// Get admin role ID
	var roleID string
	err = db.QueryRowContext(ctx, "SELECT role_id FROM roles WHERE name = ?", "ADMIN").Scan(&roleID)
	if err != nil {
		log.Fatal("Find admin role failed:", err)
	}
	fmt.Printf("Admin role ID: %s\n", roleID)

	// Assign admin role
	_, err = db.ExecContext(ctx,
		"INSERT IGNORE INTO user_roles (user_id, role_id) VALUES (?, ?)",
		userID, roleID,
	)
	if err != nil {
		log.Fatal("Assign role failed:", err)
	}
	fmt.Println("Admin role assigned successfully!")

	// Verify
	var count int
	db.QueryRowContext(ctx, "SELECT COUNT(*) FROM user_roles WHERE user_id = ? AND role_id = ?", userID, roleID).Scan(&count)
	fmt.Printf("Role assignment verified: %d rows\n", count)

	fmt.Printf("\n========================================\n")
	fmt.Printf("  ADMIN USER READY\n")
	fmt.Printf("========================================\n")
	fmt.Printf("  Email:    admin@admin.com\n")
	fmt.Printf("  Password: Adm1n123*\n")
	fmt.Printf("  Login:    https://192.168.5.106:8443/login\n")
	fmt.Printf("========================================\n")
}
