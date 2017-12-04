package models

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

func NewUser(db *sqlx.DB) *User {
	user := &User{}
	user.db = db
	user.table = "users"
	user.hasID = true

	return user
}

type UserRow struct {
	ID       int64  `db:"id"`
	Username string `db:"username"`
	Password string `db:"password"`
}

type User struct {
	Base
}

func (u *User) userRowFromSqlResult(tx *sqlx.Tx, sqlResult sql.Result) (*UserRow, error) {
	userId, err := sqlResult.LastInsertId()
	if err != nil {
		return nil, err
	}

	return u.GetById(tx, userId)
}

// AllUsers returns all user rows.
func (u *User) AllUsers(tx *sqlx.Tx) ([]*UserRow, error) {
	users := []*UserRow{}
	query := fmt.Sprintf("SELECT * FROM %v", u.table)
	err := u.db.Select(&users, query)

	return users, err
}

// GetById returns record by id.
func (u *User) GetById(tx *sqlx.Tx, id int64) (*UserRow, error) {
	user := &UserRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE id=?", u.table)
	err := u.db.Get(user, query, id)

	return user, err
}

// GetByEmail returns record by email.
func (u *User) GetByUsername(tx *sqlx.Tx, username string) (*UserRow, error) {
	user := &UserRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE username=?", u.table)
	err := u.db.Get(user, query, username)

	return user, err
}

// GetByEmail returns record by email but checks password first.
func (u *User) GetUserByUsernameAndPassword(tx *sqlx.Tx, username, password string) (*UserRow, error) {
	user, err := u.GetByUsername(tx, username)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, err
	}

	return user, err
}

// Signup create a new record of user.
func (u *User) Signup(tx *sqlx.Tx, username, password, passwordAgain string) (*UserRow, error) {
	if username == "" {
		return nil, errors.New("Username cannot be blank.")
	}
	if password == "" {
		return nil, errors.New("Password cannot be blank.")
	}
	if password != passwordAgain {
		return nil, errors.New("Password is invalid.")
	}

	//same username can't exist

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 5)
	if err != nil {
		return nil, err
	}

	data := make(map[string]interface{})
	data["username"] = username
	data["password"] = hashedPassword

	sqlResult, err := u.InsertIntoTable(tx, data)
	if err != nil {
		return nil, err
	}

	return u.userRowFromSqlResult(tx, sqlResult)
}

// UpdateEmailAndPasswordById updates user email and password.
func (u *User) UpdateUsernameAndPasswordById(tx *sqlx.Tx, userId int64, username, password, passwordAgain string) (*UserRow, error) {
	data := make(map[string]interface{})

	if username != "" {
		data["username"] = username
	}

	if password != "" && passwordAgain != "" && password == passwordAgain {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 5)
		if err != nil {
			return nil, err
		}

		data["password"] = hashedPassword
	}

	if len(data) > 0 {
		_, err := u.UpdateByID(tx, data, userId)
		if err != nil {
			return nil, err
		}
	}

	return u.GetById(tx, userId)
}
