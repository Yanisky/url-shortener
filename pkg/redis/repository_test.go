package redis

import (
	"flag"
	"log"
	"os"
	"testing"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/speps/go-hashids"
	"github.com/yanisky/url-shortener/internal/testutils"
	domain "github.com/yanisky/url-shortener/pkg"
)

var (
	testRedisCache = flag.Bool("redis", false, "run redis integration tests")
	testHasher     *hashids.HashID
	testRepo       *redisRepository
	testConn       *redis.Client
)

func TestMain(m *testing.M) {
	os.Exit(deferableTestMain(m))
}
func deferableTestMain(m *testing.M) int {
	var (
		redisURL = flag.String("test_redis_url", "", "redis url")
	)
	flag.Parse()
	timeout := 60 * time.Second
	testHasher = testutils.CreateHasherForTesting("testsalt")
	testRepo = &redisRepository{
		hasher: testHasher,
	}
	if *testRedisCache {
		conn, err := newRedisClient(*redisURL, timeout)
		if err != nil {
			log.Fatal(err)
			return 1
		}
		defer conn.Close()
		testConn = conn
		testRepo.conn = conn
	}
	return m.Run()
}

func TestCacheShouldStoreCache(t *testing.T) {
	if *testRedisCache == false {
		return
	}
	expectedURL := domain.URL{
		Hash:      "test-hash",
		Full:      "https://www.example.com",
		CreatedAt: time.Now().AddDate(0, 0, -1).UTC(),
	}
	// clean up
	defer func(conn *redis.Client, key string) {
		conn.HDel(key, "created_at", "url")
	}(testRepo.conn, expectedURL.Hash)

	if err := testRepo.Cache(expectedURL); err != nil {
		t.Fatal("Failed inserting to cache", err)
	}
	data, err := testRepo.conn.HGetAll(expectedURL.Hash).Result()
	if err != nil {
		t.Fatal("Failed to get from cache", err)
	}
	if len(data) == 0 {
		t.Fatal("Failed to set in cache")
	}
	if data["url"] != expectedURL.Full || data["created_at"] != expectedURL.CreatedAt.Format(time.RFC3339) {
		t.Fatal("Type problem", data["created_at"], data["url"], expectedURL)
	}
}

func TestCacheShouldReplaceCache(t *testing.T) {
	if *testRedisCache == false {
		return
	}
	hash := "test-hash"
	expectedURL := domain.URL{
		Hash:      hash,
		Full:      "https://www.example.com",
		CreatedAt: time.Now().AddDate(0, 0, -1).UTC(),
	}
	// clean up
	defer func(conn *redis.Client, key string) {
		conn.HDel(key, "created_at", "url")
	}(testRepo.conn, hash)

	// insert dummy
	if err := testRepo.Cache(domain.URL{Hash: hash, Full: "dummy"}); err != nil {
		t.Fatal("Failed inserting to cache", err)
	}
	// replace it
	if err := testRepo.Cache(expectedURL); err != nil {
		t.Fatal("Failed inserting to cache", err)
	}

	data, err := testRepo.conn.HGetAll(expectedURL.Hash).Result()
	if err != nil {
		t.Fatal("Failed to get from cache", err)
	}
	if len(data) == 0 {
		t.Fatal("Failed to set in cache")
	}
	if data["url"] != expectedURL.Full || data["created_at"] != expectedURL.CreatedAt.Format(time.RFC3339) {
		t.Fatal("Type problem", data["created_at"], data["url"], expectedURL)
	}
}

func TestFindShouldReturnInvalidUrlError(t *testing.T) {
	if _, err := testRepo.Find(""); err != domain.ErrorInvalidURL {
		t.Fatal("Repo should return an URL invalid error but got:", err)
	}
	if _, err := testRepo.Find(" "); err != domain.ErrorInvalidURL {
		t.Fatal("Repo should return an URL invalid error but got:", err)
	}
	if _, err := testRepo.Find("1"); err != domain.ErrorInvalidURL {
		t.Fatal("Repo should return an URL invalid error but got:", err)
	}
}

func TestFindShouldReturnUrlNotFound(t *testing.T) {
	if *testRedisCache == false {
		return
	}
	hash, _ := testHasher.EncodeInt64([]int64{8})
	if _, err := testRepo.Find(hash); err != domain.ErrorURLNotFound {
		t.Fatal("Repo should return an URL not found error but got:", err)
	}
}

func TestFindReturnsUrl(t *testing.T) {
	if *testRedisCache == false {
		return
	}
	hash, err := testHasher.EncodeInt64([]int64{8})
	if err != nil {
		t.Fatal("Failed to hash", err)
	}
	expected := domain.URL{
		Hash:      hash,
		Full:      "https://www.example.com",
		CreatedAt: time.Now().AddDate(0, 0, -1).UTC(),
	}
	// clean up
	defer func(conn *redis.Client, key string) {
		conn.HDel(key, "created_at", "url")
	}(testRepo.conn, hash)

	if err = testRepo.Cache(expected); err != nil {
		t.Fatal("Failed to insert to cache", err)
	}

	actual, err := testRepo.Find(expected.Hash)
	if err != nil {
		t.Fatal("Failed to find from cache", err)
	}
	expectedDate := expected.CreatedAt.UTC().Format(time.RFC3339)
	actualDate := actual.CreatedAt.UTC().Format(time.RFC3339)
	if actual.Hash != expected.Hash || actual.Full != expected.Full || actualDate != expectedDate {
		t.Fatal("Structs dont match", expected, actual)
	}
}
