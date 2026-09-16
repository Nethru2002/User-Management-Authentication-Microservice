package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
)

func main() {
	ctx := context.Background()
	dsn := "postgres://postgres:nethru2002@localhost:5432/postgres?sslmode=disable"

	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer conn.Close(ctx)

	_, err = conn.Exec(ctx, "CREATE DATABASE enterprise_db;")
	if err != nil {
		fmt.Println("Database might already exist, skipping creation:", err)
	} else {
		fmt.Println("Database 'enterprise_db' created successfully!")
	}

	dsnDB := "postgres://postgres:nethru2002@localhost:5432/enterprise_db?sslmode=disable"
	connDB, err := pgx.Connect(ctx, dsnDB)
	if err != nil {
		log.Fatalf("Unable to connect to enterprise_db: %v", err)
	}
	defer connDB.Close(ctx)

	query := `
	CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY,
		email VARCHAR(255) UNIQUE NOT NULL,
		password VARCHAR(255) NOT NULL,
		role VARCHAR(50) NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE NOT NULL,
		updated_at TIMESTAMP WITH TIME ZONE NOT NULL
	);`

	_, err = connDB.Exec(ctx, query)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	fmt.Println("Table 'users' migrated successfully!")
}