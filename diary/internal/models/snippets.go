package models

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"time"
)

type Snippet struct {
	ID      int
	Title   string
	Content string
	Created time.Time
	Expires time.Time
}

type SnippetModel struct {
	DB *pgxpool.Pool
}

func (sm *SnippetModel) Insert(title string, content string, expires int) (int, error) {
	conn, err := sm.DB.Acquire(context.Background())
	if err != nil {
		fmt.Printf("Unable to acquire a database connection: %v\n", err)
		return 0, nil
	}
	defer conn.Release()

	var stmt string
	if expires == 7 {
		stmt = "INSERT INTO snippets (title, content, created, expires) VALUES ($1, $2, current_timestamp, current_timestamp + interval '7 days') RETURNING id"
	} else if expires == 1 {
		stmt = "INSERT INTO snippets (title, content, created, expires) VALUES ($1, $2, current_timestamp, current_timestamp + interval '1 day') RETURNING id"
	} else {
		stmt = "INSERT INTO snippets (title, content, created, expires) VALUES ($1, $2, current_timestamp, current_timestamp + interval '1 year') RETURNING id"
	}

	row := conn.QueryRow(context.Background(),
		stmt,
		title, content)
	var id int
	err = row.Scan(&id)
	if err != nil {
		fmt.Printf("Unable to INSERT: %v\n", err)
		return 0, nil
	}

	return id, err
}

func (sm *SnippetModel) Get(snippetId int) (*Snippet, error) {
	conn, err := sm.DB.Acquire(context.Background())
	if err != nil {
		fmt.Printf("Unable to acquire a database connection: %v\n", err)
		return nil, nil
	}
	defer conn.Release()

	row := conn.QueryRow(context.Background(),
		"SELECT id, title, content, created, expires FROM snippets WHERE expires > current_timestamp AND id = $1",
		snippetId)
	s := &Snippet{}
	err = row.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
	if err != nil {

		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}
	return s, err
}

func (sm *SnippetModel) Latest() ([]*Snippet, error) {
	conn, err := sm.DB.Acquire(context.Background())
	if err != nil {
		fmt.Printf("Unable to acquire a database connection: %v\n", err)
		return nil, nil
	}
	defer conn.Release()

	rows, err := conn.Query(context.Background(),
		"SELECT id, title, content, created, expires FROM snippets WHERE expires > current_timestamp ORDER BY id DESC LIMIT 10",
	)
	snippets := []*Snippet{}
	for rows.Next() {
		s := &Snippet{}
		err = rows.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
		if err != nil {
			return nil, err
		}
		snippets = append(snippets, s)
	}
	if err != nil {

		if errors.Is(err, pgx.ErrNoRows) {
			fmt.Println("ErrNoRecord from snippets.go")
			return nil, ErrNoRecord
		} else {
			fmt.Println("Another error from snippets.go")

			return nil, err
		}
	}
	return snippets, err
}
