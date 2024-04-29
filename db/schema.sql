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
CREATE TABLE IF NOT EXISTS inbox (
    message_id INT NOT NULL,
    receiver_id INT NOT NULL,
    FOREIGN KEY (message_id) REFERENCES messages(message_id),
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

