package postgresql

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/speps/go-hashids"
	domain "github.com/yanisky/url-shortener/pkg"
)

type postgreSQLRepository struct {
	conn    *pgxpool.Pool
	timeout time.Duration
	hasher  *hashids.HashID
}

func NewPostgreSQLRepository(dbURL string, timeout time.Duration, hasher *hashids.HashID) (*postgreSQLRepository, error) {

	repo := &postgreSQLRepository{
		timeout: timeout,
		hasher:  hasher,
	}
	conn, err := createConnectionPool(dbURL, repo.timeout)
	if err != nil {
		return nil, err
	}
	repo.conn = conn
	return repo, nil
}

func (r *postgreSQLRepository) Find(urlHash string) (domain.URL, error) {
	if len(urlHash) == 0 {
		return domain.URL{}, domain.ErrorInvalidURL
	}
	ids, err := r.hasher.DecodeInt64WithError(urlHash)
	if err != nil {
		return domain.URL{}, domain.ErrorInvalidURL
	}
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	dbUrl := &domain.URL{}
	err = r.conn.QueryRow(ctx, "SELECT url, short, created_at FROM urls WHERE id=$1", ids[0]).Scan(
		&dbUrl.Full,
		&dbUrl.Hash,
		&dbUrl.CreatedAt,
	)
	if err != nil {
		if err.Error() == pgx.ErrNoRows.Error() {
			return domain.URL{}, domain.ErrorURLNotFound
		}
		return domain.URL{}, err
	}

	return *dbUrl, nil
}

//Cache doesn't do anything because we use table as "cache"
func (r *postgreSQLRepository) Cache(url domain.URL) error {
	return nil
}

func (r *postgreSQLRepository) Create(url string) (domain.URL, error) {
	var id int64
	returnURL := domain.URL{}
	fullURL, err := domain.NormalizeURL(url)
	if err != nil {
		return returnURL, domain.ErrorInvalidURL
	}
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	err = r.conn.QueryRow(
		ctx,
		"INSERT INTO urls (short, url) VALUES ($1, $2) RETURNING id, created_at",
		time.Now().String(),
		fullURL,
	).Scan(&id, &returnURL.CreatedAt)
	if err != nil {
		return returnURL, err
	}
	hash, err := r.hasher.EncodeInt64([]int64{id})
	if err != nil {
		return returnURL, err
	}
	_, err = r.conn.Exec(context.Background(), "UPDATE urls SET short=$1 WHERE id=$2", hash, id)
	if err != nil {
		return returnURL, err
	}
	returnURL.Hash = hash
	returnURL.Full = fullURL

	return returnURL, nil
}

func (r *postgreSQLRepository) CreateURLView(urlHash string) error {
	if len(urlHash) == 0 {
		return domain.ErrorInvalidURL
	}

	ids, err := r.hasher.DecodeInt64WithError(urlHash)
	if err != nil {
		return domain.ErrorInvalidURL
	}
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	_, err = r.conn.Exec(ctx, "INSERT INTO url_views (url_id) VALUES ($1)", ids[0])
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func (r *postgreSQLRepository) Stats(urlHash string) (domain.URLViewStats, error) {
	if len(urlHash) == 0 {
		return domain.URLViewStats{}, domain.ErrorInvalidURL
	}

	ids, err := r.hasher.DecodeInt64WithError(urlHash)
	if err != nil {
		return domain.URLViewStats{}, domain.ErrorInvalidURL
	}

	var count, pastDayCount, pastWeekCount int

	// very inefficient like a vintage car.
	countSQL := "SELECT Count(*) FROM url_views WHERE url_id=$1"
	pastWeekSQL := "SELECT Count(*) FROM url_views WHERE url_id=$1 AND created_at >= NOW() - interval '1 week'"
	pastDaySQL := "SELECT Count(*) FROM url_views WHERE url_id=$1 AND created_at >= NOW() - interval '1 day'"

	err = r.conn.QueryRow(context.Background(), countSQL, ids[0]).Scan(&count)
	if err != nil {
		fmt.Println(err)
		return domain.URLViewStats{}, err
	}

	err = r.conn.QueryRow(context.Background(), pastWeekSQL, ids[0]).Scan(&pastWeekCount)
	if err != nil {
		fmt.Println(err)
		return domain.URLViewStats{}, err
	}

	err = r.conn.QueryRow(context.Background(), pastDaySQL, ids[0]).Scan(&pastDayCount)
	if err != nil {
		fmt.Println(err)
		return domain.URLViewStats{}, err
	}

	return domain.URLViewStats{
		Count:         count,
		PastWeekCount: pastWeekCount,
		PastDayCount:  pastDayCount,
	}, nil
}

func createConnectionPool(database string, timeout time.Duration) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(database)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	conn, err := pgxpool.ConnectConfig(ctx, poolConfig)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
