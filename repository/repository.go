package repository

import (
	"database/sql"
	"errors"
	"log"
)

type Repository interface {
	Create(client Client) error
	Get(clientID string) (*Client, error)
	Update(client Client) error
	Delete(clientID string) error
}

type postgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) Repository {
	r := postgresRepository{db: db}
	r.InitSchema()
	return &r
}

func (r *postgresRepository) InitSchema() {
	_, err := r.db.Exec(`
	CREATE TABLE IF NOT EXISTS clients (
    id TEXT PRIMARY KEY,
    capacity INTEGER NOT NULL,
    rate INTEGER NOT NULL
);`)
	if err != nil {
		log.Println("Schema did not init")
	}
}

func (r *postgresRepository) Create(client Client) error {
	_, err := r.db.Exec(`
	INSERT INTO clients (id, capacity, rate)
	VALUES($1, $2, $3)
	ON CONFLICT (id) DO NOTHING
	`, client.ID, client.Capacity, client.Rate)
	if err != nil {
		return err
	}
	return nil
}

func (r *postgresRepository) Update(client Client) error {
	_, err := r.db.Exec(`
	UPDATE clients
	SET 
	    capacity = $1, rate = $2
	WHERE id = $3
	`, client.Capacity, client.Rate, client.ID)
	if err != nil {
		return err
	}
	return nil
}

func (r *postgresRepository) Delete(id string) error {
	_, err := r.db.Exec(`
	DELETE FROM clients
	WHERE id = $1
	`, id)
	if err != nil {
		return err
	}
	return nil
}

func (r *postgresRepository) Get(id string) (*Client, error) {
	var client Client
	err := r.db.QueryRow(`
	SELECT
	    id, rate, capacity
	FROM clients
	WHERE id = $1
	`, id).Scan(&client.ID, &client.Rate, &client.Capacity)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &client, nil
}
