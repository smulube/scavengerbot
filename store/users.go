package store

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	null "gopkg.in/guregu/null.v4"
)

// User type
type User struct {
	ID        int64       `json:"id" db:"id"`
	UserName  null.String `json:"userName" db:"user_name"`
	FirstName string      `json:"firstName" db:"first_name"`
	Admin     bool        `json:"admin" db:"admin"`
	TeamID    null.Int    `json:"teamId" db:"team_id"`
}

// GetUser returns the user identified by the given integer id
func GetUser(tx *sqlx.Tx, id int) (*User, error) {
	var user User

	err := tx.Get(&user, "SELECT * FROM users WHERE id=$1", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("Unexpected error: %v", err)
	}

	return &user, nil
}

// SaveUser attempts to save the specified user
func SaveUser(tx *sqlx.Tx, user *User) error {
	query := `INSERT INTO users
		(id, user_name, first_name, team_id, admin)
	VALUES
		(:id, :user_name, :first_name, :team_id, :admin)
	ON CONFLICT (id)
	DO UPDATE SET user_name = EXCLUDED.user_name,
	first_name = EXCLUDED.first_name,
	team_id = EXCLUDED.team_id,
	admin = EXCLUDED.admin`

	_, err := tx.NamedExec(query, user)
	if err != nil {
		return err
	}

	return nil
}
