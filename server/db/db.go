package db

import (
	"database/sql"
	"log"
	"sort"
	"strconv"

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

func MustInitDB(connectionString string) *DB {
	db, err := InitDB(connectionString)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

type DBErr struct {
	Err string
	Msg string
}

func newdbeErr(e string) *DBErr {
	return &DBErr{Err: e}
}

func newdbeErrAndMsg(e, m string) *DBErr {
	return &DBErr{Err: e, Msg: m}
}

func (dbe *DBErr) Error() string {
	return dbe.Err
}

var (
	UserExists               = newdbeErr("user name already exists")
	UserDoesNotExist         = newdbeErr("user name does not exist")
	UserPasswordDoesNotMatch = newdbeErr("user password does not match")
	InvalidUserID            = newdbeErr("user password does not match")
)

func (d *DB) Close() error {
	return d.db.Close()
}

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
	if err := d.db.QueryRow("SELECT user_id, name, password FROM users WHERE user_name = $1", userName).Scan(&u.Id, &u.Name, &actualPass); err != nil {
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

func (d *DB) SaveMsg(m *Message) (id int, err error) {
	err = d.db.QueryRow(insertMessageQuery, m.SenderId, m.ReceiverId, m.MessageText).Scan(&id)
	return
}

// either a stirng or an int :)
func (d *DB) FindUserById(userId interface{}) (u User, err error) {
	switch userId.(type) {
	case int:
		break
	case string:
		v, _ := userId.(string)
		_, err := strconv.Atoi(v)
		if err != nil {
			return u, InvalidUserID
		}
	default:
		return u, InvalidUserID
	}

	err = d.db.QueryRow("SELECT user_id, name, user_name FROM users WHERE user_id = $1", userId).Scan(&u.Id, &u.Name, &u.UserName)
	if err == sql.ErrNoRows {
		return User{}, UserDoesNotExist
	}
	return
}

func (d *DB) FindUserByUserName(userName string) (u User, err error) {
	err = d.db.QueryRow("SELECT user_id, name FROM users WHERE user_name = $1", userName).Scan(&u.Id, &u.Name)
	if err == sql.ErrNoRows {
		return User{}, UserDoesNotExist
	}
	return
}

// sorted by last send but the actyual messages are in reverse
func (d *DB) GetAllMessagesOfUser(userID int) ([]*Inbox, error) {
	user, err := d.FindUserById(userID)
	if err != nil {
		return nil, err
	}

	const q = `SELECT message_id, sender_id, receiver_id, message_text, sent_at
		FROM messages
		WHERE receiver_id = $1 OR sender_id = $1`

	rows, err := d.db.Query(q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	inbox := map[int]*Inbox{}
	for rows.Next() {
		m := Message{}
		err = rows.Scan(&m.Id, &m.SenderId, &m.ReceiverId, &m.MessageText, &m.SentAt)
		if err != nil {
			return nil, err
		}
		uId := m.SenderId
		if user.Id == m.SenderId {
			uId = m.ReceiverId
		}

		w, ok := inbox[uId]
		if ok {
			w.Messages = append(w.Messages, m)
		} else {
			u, err := d.FindUserById(uId)
			if err != nil {
				log.Printf("this should not happend err: %v\n", err)
				return nil, err
			}
			inbox[uId] = &Inbox{u, []Message{m}}
		}
	}

	a := []*Inbox{}

	for _, v := range inbox {
		a = append(a, v)
	}

	sort.Slice(a, func(i, j int) bool {
		return a[i].Messages[len(a[i].Messages)-1].SentAt.Unix() > a[j].Messages[len(a[j].Messages)-1].SentAt.Unix()
	})

	return a, nil
}
