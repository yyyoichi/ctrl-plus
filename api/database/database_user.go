package database

import (
	"database/sql"
	"errors"
	"himakiwa/auth"
	"time"
)

var (
	ErrInValidParams   = errors.New("invalid params check your params")
	ErrInvalidEmail    = errors.New("invalid email does not exist this.email")
	ErrInvalidPassword = errors.New("invalid password does not match password")
)

type SignInUser struct {
	Name     string
	Email    string
	Password string
}

// ユーザ登録
func (u *SignInUser) SignIn(db *sql.DB) (int, error) {
	// validation
	if u.Email == "" || u.Password == "" {
		return 0, ErrInValidParams
	}

	// hashed password
	hashedPassword, err := auth.HashPassword(u.Password)
	if err != nil {
		return 0, err
	}

	// connect db
	if db == nil {
		var err error
		db, err = GetDatabase()
		if err != nil {
			return 0, err
		}
		defer db.Close()
	}

	// bigin transaction and insert db
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}

	// exec insert
	s := `INSERT INTO users (name, email, password, create_at, login_at, update_at) VALUES(?, ?, ?, ?, ?, ?)`
	now := time.Now()
	result, err := tx.Exec(s, u.Name, u.Email, hashedPassword, now, now, now)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	// get id
	id64, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	tx.Commit()

	return int(id64), nil
}

// emailの重複調査
func ExistEmail(db *sql.DB, email string) (bool, error) {
	// connect db
	if db == nil {
		var err error
		db, err = GetDatabase()
		if err != nil {
			return true, err
		}
		defer db.Close()
	}

	// count id
	var i int64
	s := `SELECT count(id) FROM users WHERE email = ?`
	err := db.QueryRow(s, email).Scan(&i)
	if err != nil {
		return true, err
	}

	return 0 < i, nil
}

// ユーザ行そのものを消す
func DeleteUserRow(db *sql.DB, userId int) error {
	// validate
	if userId == 0 {
		return ErrInValidParams
	}
	// connect db
	if db == nil {
		var err error
		db, err = GetDatabase()
		if err != nil {
			return err
		}
		defer db.Close()
	}

	// exec delete row
	s := `DELETE FROM users WHERE id = ?`
	_, err := db.Exec(s, userId)
	if err != nil {
		return err
	}
	return nil
}

// ユーザの削除
func DeleteUser(db *sql.DB, userId int) error {
	// validate
	if userId == 0 {
		return ErrInValidParams
	}
	// connect db
	if db == nil {
		var err error
		db, err = GetDatabase()
		if err != nil {
			return err
		}
		defer db.Close()
	}

	// deleted flag on
	s := `UPDATE user SET deleted=1 WHERE id = ?`
	_, err := db.Exec(s, userId)
	if err != nil {
		return err
	}
	return nil
}

type DatabaseUser struct {
	Id       int
	Name     string
	Email    string
	LoginAt  time.Time
	CreateAt time.Time
	UpdateAt time.Time
	Deleted  bool
	// db get
	dbDeleted int64
}

// ユーザ照会
func QueryUser(db *sql.DB, userId int) (*DatabaseUser, error) {
	// validate
	if userId == 0 {
		return nil, ErrInValidParams
	}
	// connect db
	if db == nil {
		var err error
		db, err = GetDatabase()
		if err != nil {
			return nil, err
		}
		defer db.Close()
	}

	// query user
	du := &DatabaseUser{Id: userId}
	s := `SELECT name, email, login_at, create_at, update_at, deleted FROM users WHERE id = ?`
	row := db.QueryRow(s, userId)
	err := row.Scan(&du.Name, &du.Email, &du.LoginAt, &du.CreateAt, &du.UpdateAt, &du.dbDeleted)
	if err != nil {
		return nil, err
	}

	// map
	du.Deleted = du.dbDeleted == 1
	return du, nil
}

type LoginUser struct {
	Email    string
	Password string
}

func (u *LoginUser) Login(db *sql.DB) (*DatabaseUser, error) {
	// validate
	if u.Email == "" || u.Password == "" {
		return nil, ErrInValidParams
	}
	// connect db
	if db == nil {
		var err error
		db, err = GetDatabase()
		if err != nil {
			return nil, err
		}
		defer db.Close()
	}

	// query user
	var hashedPassword string
	du := &DatabaseUser{Email: u.Email}
	s := `SELECT id, name, email, password, login_at, create_at, update_at, deleted FROM users WHERE email = ?`
	row := db.QueryRow(s, u.Email)
	err := row.Scan(&du.Id, &du.Name, &du.Email, &hashedPassword, &du.LoginAt, &du.CreateAt, &du.UpdateAt, &du.dbDeleted)
	if err != nil {
		return nil, ErrInvalidEmail
	}

	// map
	du.Deleted = du.dbDeleted == 1
	if du.Deleted {
		return nil, ErrInValidParams
	}

	// check password
	result, err := auth.ComparePasswordAndHash(u.Password, hashedPassword)
	if err != nil {
		return nil, err
	}
	if !result {
		return nil, ErrInvalidPassword
	}

	return du, nil
}