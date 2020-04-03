package redis

import (
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/speps/go-hashids"
	domain "github.com/yanisky/url-shortener/pkg"
)

type redisRepository struct {
	conn   *redis.Client
	hasher *hashids.HashID
}

func (r *redisRepository) Find(urlHash string) (domain.URL, error) {
	if len(urlHash) == 0 {
		return domain.URL{}, domain.ErrorInvalidURL
	}
	_, err := r.hasher.DecodeInt64WithError(urlHash)
	if err != nil {
		return domain.URL{}, domain.ErrorInvalidURL
	}
	data, err := r.conn.HGetAll(urlHash).Result()
	if err != nil {
		return domain.URL{}, err
	}
	if len(data) == 0 {
		return domain.URL{}, domain.ErrorURLNotFound
	}
	createdAt, err := time.Parse(time.RFC3339, data["created_at"])
	if err != nil {
		// log error, but ignore, it's not important enough to fail a return
		createdAt = time.Now().UTC()
	}
	return domain.URL{
		Hash:      urlHash,
		Full:      data["url"],
		CreatedAt: createdAt,
	}, nil
}

func (r *redisRepository) Cache(url domain.URL) error {
	data := map[string]interface{}{
		"url":        url.Full,
		"created_at": url.CreatedAt.UTC(),
	}
	_, err := r.conn.HSet(url.Hash, data).Result()
	if err != nil {
		return err
	}

	return nil
}

func NewRedisRepository(redisURL string, timeout time.Duration, hasher *hashids.HashID) (*redisRepository, error) {
	repo := &redisRepository{
		hasher: hasher,
	}
	client, err := newRedisClient(redisURL, timeout)
	if err != nil {
		return nil, err
	}
	repo.conn = client

	return repo, nil
}

func newRedisClient(redisURL string, redisTimeout time.Duration) (*redis.Client, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}
	opts.ReadTimeout = redisTimeout
	opts.WriteTimeout = redisTimeout

	client := redis.NewClient(opts)
	_, err = client.Ping().Result()

	if err != nil {
		return nil, err
	}
	return client, nil
}
