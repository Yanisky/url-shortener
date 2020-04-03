package testutils

import (
	"context"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

func TruncateAllTables(db *pgxpool.Pool) error {
	sql := "TRUNCATE TABLE urls, url_views"
	if _, err := db.Exec(context.Background(), sql); err != nil {
		return err
	}
	return nil
}
func TruncateUrlsTable(db *pgxpool.Pool) error {
	sql := "TRUNCATE TABLE urls"
	if _, err := db.Exec(context.Background(), sql); err != nil {
		return err
	}
	return nil
}
func TruncateUrlViewsTable(db *pgxpool.Pool) error {
	sql := "TRUNCATE TABLE url_views"
	if _, err := db.Exec(context.Background(), sql); err != nil {
		return err
	}
	return nil
}
func InsertUrl(db *pgxpool.Pool, full string) (int64, error) {
	sql := "INSERT INTO urls (url, short) VALUES ($1, $2) RETURNING id"
	var id int64
	if err := db.QueryRow(context.Background(), sql, full, time.Now().String()).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil

}

func GeUrlId(db *pgxpool.Pool, full string, hash string) (int64, error) {
	sql := "SELECT id FROM urls WHERE url=$1 AND short=$2"
	var id int64
	if err := db.QueryRow(context.Background(), sql, full, hash).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

type URLView struct {
	URLId     int64
	CreatedAt time.Time
}

func GetUrlView(db *pgxpool.Pool, id int64) (*URLView, error) {
	sql := "SELECT url_id, created_at FROM url_views WHERE url_id=$1"
	view := URLView{}
	if err := db.QueryRow(context.Background(), sql, id).Scan(&view.URLId, &view.CreatedAt); err != nil {
		return nil, err
	}
	return &view, nil
}

func InsertUrlView(db *pgxpool.Pool, id int64, createdAt time.Time) error {
	sql := "INSERT INTO url_views (url_id, created_at) VALUES ($1, $2)"
	if _, err := db.Exec(context.Background(), sql, id, createdAt); err != nil {
		return err
	}
	return nil

}
