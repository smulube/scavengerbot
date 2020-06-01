package store

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// Team type
type Team struct {
	ID   int64  `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
}

// GetTeams returns a slice of all teams in the system
func GetTeams(tx *sqlx.Tx) ([]*Team, error) {
	query := `SELECT * FROM teams ORDER BY name`

	rows, err := tx.Queryx(query)
	if err != nil {
		return nil, fmt.Errorf("Unexpected error querying for teams: %v", err)
	}

	teams := []*Team{}

	for rows.Next() {
		var t Team
		err = rows.StructScan(&t)
		if err != nil {
			return nil, fmt.Errorf("Unexpected error scanning Team struct: %v", err)
		}
		teams = append(teams, &t)
	}

	return teams, nil
}

// CreateTeam inserts a new team into the DB
func CreateTeam(tx *sqlx.Tx, team *Team) error {
	query := `INSERT INTO teams (name) VALUES (:name)`

	_, err := tx.NamedExec(query, team)
	if err != nil {
		return fmt.Errorf("Unable to create team, team name may already exist: %v", err)
	}

	return nil
}

// GetTeamByID returns a team given its numerical ID
func GetTeamByID(tx *sqlx.Tx, id int) (*Team, error) {
	query := `SELECT * FROM teams WHERE id=$1`

	var t Team

	err := tx.Get(&t, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("Unexpected error: %v", err)
	}

	return &t, nil
}

// GetTeamByName returns a team given its name
func GetTeamByName(tx *sqlx.Tx, name string) (*Team, error) {
	query := `SELECT * FROM teams WHERE name=$1`

	var t Team

	err := tx.Get(&t, query, name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("Unexpected error: %v", err)
	}

	return &t, nil
}
