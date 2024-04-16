package db

import (
	"database/sql"
	"errors"

	_ "github.com/lib/pq"
)

type DB struct{ db *sql.DB }

// opens the connectin and creates the table if not exits
//
// connectionString = "user=postgres password=pass host=localhost port=5432 sslmode=disable"
func InitDB(connectionString string) (*DB, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(schema)
	if err != nil {
		return nil, err
	}

	return &DB{db: db}, nil
}

var (
	UserExists               = errors.New("user name already exists")
	UserDoesNotExist         = errors.New("user name does not exist")
	UserPasswordDoesNotMatch = errors.New("user password does not match")
)

func (d *DB) Singup(name, userName, password string) (User, error) {
	var count int
	if err := d.db.QueryRow("SELECT COUNT(*) FROM users WHERE user_name = $1", userName).Scan(&count); err != nil {
		return User{}, err
	} else if count > 0 {
		return User{}, UserExists
	}

	var id int
	if err := d.db.QueryRow(insertUserQuery, name, userName, password).Scan(&id); err != nil {
		return User{}, err
	}

	return User{
		Id:       id,
		Name:     name,
		UserName: userName,
	}, nil
}

func (d *DB) Singin(userName, password string) (User, error) {
	u := User{}
	var actualPass string
	if err := d.db.QueryRow("SELECT id, name, password FROM users WHERE user_name = $1", userName).Scan(&u.Id, &u.Name, &actualPass); err != nil {
		if err == sql.ErrNoRows {
			return User{}, UserDoesNotExist
		}
		return User{}, err
	}

	if password != actualPass {
		return User{}, UserPasswordDoesNotMatch
	}

	u.UserName = userName
	return u, nil
}

func (d *DB) SaveMsg(senderId, receiverId, msg string) (id int, err error) {
	err = d.db.QueryRow(insertMessageQuery, senderId, receiverId, msg).Scan(&id)
	return
}

func (d *DB) FindUser(userName string) (u User, err error) {
	err = d.db.QueryRow("SELECT id, name FROM users WHERE user_name = $1", userName).Scan(&u.Id, &u.Name)
	if err == sql.ErrNoRows {
		return User{}, UserDoesNotExist
	}
	return
}

func (d *DB) GetInbox(userID string) ([]User, error) {
	const q = `SELECT u.user_id, u.name
        FROM users u
        JOIN (
            SELECT DISTINCT sender_id AS user_id
            FROM messages
            WHERE receiver_id = $1 OR sender_id = $1
        ) m ON u.user_id = m.user_id`

	rows, err := d.db.Query(q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []User{}
	for rows.Next() {
		u := User{}
		err = rows.Scan(&u.UserName, &u.Name)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, nil
}
