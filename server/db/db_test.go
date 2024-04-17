package db

import (
	"database/sql"
	"fmt"
	"log"
	"testing"
)

func TestMain(t *testing.T) {
	d, err := InitDB("user=postgres password=pass host=localhost port=5432 sslmode=disable") // dbname=your_database
	if err != nil {
		t.Fatal(err)
	}

	db := d.db

	err = dropAll(db)
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec(schema)
	if err != nil {
		t.Fatal(err)
	}

	// Insert demo data into the users table
	insertUserQuery := `
		INSERT INTO users (name, email, phone, user_name, password)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING user_id
	`
	var userID1 int
	err = db.QueryRow(insertUserQuery, "John Doe", "john@example.com", "+1234567890", "johndoe", "password123").Scan(&userID1)
	if err != nil {
		log.Fatalf("Error inserting user: %v\n", err)
	}
	fmt.Printf("Inserted user with ID: %d\n", userID1)

	var userID2 int
	err = db.QueryRow(insertUserQuery, "John Doe1", "john@example2.com", "+1234467890", "lohndoe", "password123").Scan(&userID2)
	if err != nil {
		log.Fatalf("Error inserting user: %v\n", err)
	}
	fmt.Printf("Inserted user with ID: %d\n", userID2)

	// Insert demo data into the messages table
	insertMessageQuery := `
		INSERT INTO messages (sender_id, receiver_id, message_text)
		VALUES ($1, $2, $3)
		RETURNING message_id
	`
	var messageID int
	err = db.QueryRow(insertMessageQuery, userID2, userID1, "Hello, how are you?").Scan(&messageID)
	// err = db.QueryRow(insertMessageQuery, userID1, userID2, "Hello, how are you?").Scan(&messageID)
	if err != nil {
		log.Fatalf("Error inserting message: %v\n", err)
	}
	fmt.Printf("Inserted message with ID: %d\n", messageID)

	insertInboxQuery := `
		INSERT INTO inbox (sender_id, receiver_id)
		VALUES ($1, $2)
	`
	_, err = db.Exec(insertInboxQuery, userID1, userID2)
	if err != nil {
		log.Fatalf("Error inserting to inbox: %v\n", err)
	}
	fmt.Printf("Inserted message with ID: %d\n", messageID)

	fmt.Printf("user 1: ")
	fmt.Println(d.GetAllMessagesByUserId("1"))
	fmt.Printf("user 2: ")
	fmt.Println(d.GetAllMessagesByUserId("2"))
	fmt.Println(d.GetAllMessagesByUserId("4"))
}

func dropAll(db *sql.DB) error {
	_, err := db.Exec("DROP TABLE IF EXISTS users CASCADE")
	if err != nil {
		return nil
	}
	// log.Println("Users table dropped successfully")

	// Drop the messages table if it exists
	_, err = db.Exec("DROP TABLE IF EXISTS messages CASCADE")
	if err != nil {
		return nil
	}
	// log.Println("Messages table dropped successfully")

	_, err = db.Exec("DROP TABLE IF EXISTS inbox CASCADE")
	if err != nil {
		return nil
	}
	// log.Println("Inbox table dropped successfully")
	return nil
}
