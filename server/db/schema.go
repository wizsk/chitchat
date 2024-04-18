package db

import "time"

const (
	insertUserQuery string = `
		-- INSERT INTO users (name, email, phone, user_name, password)
		-- VALUES ($1, $2, $3, $4, $5)
		INSERT INTO users (name, user_name, password)
		VALUES ($1, $2, $3)
		RETURNING user_id`

	insertMessageQuery string = `
		INSERT INTO messages (sender_id, receiver_id, message_text)
		VALUES ($1, $2, $3)
		RETURNING message_id
	`

	insertInboxQuery string = `
		INSERT INTO inbox (sender_id, receiver_id)
		VALUES ($1, $2)
	`
)

type User struct {
	Id       int    `json:"id,omitemtpy"`
	Name     string `json:"name"`
	UserName string `json:"user_name"`
	Email    string `json:"email,omitempty"`
	Phone    string `json:"phone,omitempty"`
	// Password string `json:"password"`
}

type Message struct {
	Id          int       `json:"id,omitempty"`
	SenderId    int       `json:"sender_id"`
	ReceiverId  int       `json:"receiver_id"`
	MessageText string    `json:"message_text"`
	SentAt      time.Time `json:"sent_at"`
}

type AllInbox struct {
	Inbox []Inbox
}

type Inbox struct {
	User     User      `json:"user"`
	Messages []Message `json:"messages"`
}

const schema string = `
-- Create users table if not exists
CREATE TABLE IF NOT EXISTS users (
    user_id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) UNIQUE,
    phone VARCHAR(20) UNIQUE,
    user_name VARCHAR(50) UNIQUE,
    password VARCHAR(100) NOT NULL
);

-- Create messages table if not exists
CREATE TABLE IF NOT EXISTS messages (
    message_id SERIAL PRIMARY KEY,
    sender_id INT NOT NULL,
    receiver_id INT NOT NULL,
    message_text TEXT NOT NULL,
    sent_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (sender_id) REFERENCES users(user_id),
    FOREIGN KEY (receiver_id) REFERENCES users(user_id)
);

-- Create inbox table if not exists
-- for now this will be ignored
CREATE TABLE IF NOT EXISTS inbox (
    sender_id INT NOT NULL,
    receiver_id INT NOT NULL,
    FOREIGN KEY (sender_id) REFERENCES users(user_id),
    FOREIGN KEY (receiver_id) REFERENCES users(user_id)
);

-- Create user relationships table if not exists (for example, friendships)
CREATE TABLE IF NOT EXISTS user_relationships (
    user1_id INT NOT NULL,
    user2_id INT NOT NULL,
    relationship_type VARCHAR(20) NOT NULL,
    PRIMARY KEY (user1_id, user2_id),
    FOREIGN KEY (user1_id) REFERENCES users(user_id),
    FOREIGN KEY (user2_id) REFERENCES users(user_id)
);
`
