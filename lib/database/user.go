package database

import (
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

type User struct {
	Username    string    `json:"username"`
	Password    string    `json:"-"`
	Email       string    `json:"email"`
	Apikey      string    `json:"apikey,omitempty"`
	SecretToken string    `json:"secret_token,omitempty"`
	IsDeleted   bool      `json:"is_deleted"`
	IsActive    bool      `json:"is_active"`
	IsSuperuser bool      `json:"is_superuser"`
	CreatedAt   time.Time `json:"created_at,string"`
	UpdatedAt   time.Time `json:"updated_at,string"`
	db          *Database `json:"-"`
}

// SetEmail sets user email
func (self *User) SetEmail(email string) error {
	self.Email = email
	return self.Update()
}

// SetPassword sets password
func (self *User) SetPassword(password string) error {
	self.Password = password
	return self.Update()
}

// Delete deletes user
func (self *User) Delete() error {
	self.IsDeleted = true
	return self.Update()
}

// Activate deletes user
func (self *User) Activate() error {
	self.IsActive = true
	return self.Update()
}

// Deactivate deletes user
func (self *User) Deactivate() error {
	self.IsActive = false
	return self.Update()
}

// Update updates user data in database
func (self *User) Update() error {
	return self.db.Insert(`
		UPDATE users
			SET
				email=$1,
				password=$2,
				is_deleted=$3,
				is_active=$4
			WHERE username=$5;`, self.Email, self.Password, self.IsDeleted, self.IsActive, self.Username)
}

func (self *User) Marshal() (string, error) {
	b, err := json.Marshal(self)
	if nil != err {
		return "", err
	}
	return string(b), err
}

func (self *User) Unmarshal(data string) error {
	return json.Unmarshal([]byte(data), self)
}

// IsPassword checks if provided password/hash matches database record
func (self *User) IsPassword(password string) (bool, error) {
	match := false
	return match, self.db.Exec(func(conn *sql.DB) error {
		rows, err := conn.Query(`SELECT is_password($1, $2);`, self.Username, password)

		if nil != err {
			return err
		}

		for rows.Next() {
			rows.Scan(&match)
			return nil
		}

		return errors.New("Not found")
	})
}

/**
 * Social Accounts
 */
// CreateSocialAccountIfNotExists
// https://stackoverflow.com/questions/4069718/postgres-insert-if-does-not-exist-already
// ON CONFLICT DO NOTHING/UPDATE
// http://www.postgresqltutorial.com/postgresql-upsert/
func (self *User) CreateSocialAccountIfNotExists(user_id, username, account_type string) error {
	err := self.db.Insert(`
		INSERT INTO social_accounts(id, name, type, email)
			VALUES ($1, $2, $3, $4)
				ON CONFLICT DO NOTHING;
	`, user_id, username, account_type, self.Username)
	if nil != err && strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
		return nil
	}
	return nil
}
