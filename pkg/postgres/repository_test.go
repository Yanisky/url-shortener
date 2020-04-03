package postgresql

import (
	"flag"
	"log"
	"os"
	"testing"
	"time"

	"github.com/speps/go-hashids"
	"github.com/yanisky/url-shortener/internal/testutils"
	domain "github.com/yanisky/url-shortener/pkg"
)

var (
	testPostgreSQL = flag.Bool("postgres", false, "run database integration tests")
	testHasher     *hashids.HashID
	testRepo       *postgreSQLRepository
)

func TestMain(m *testing.M) {
	os.Exit(deferableTestMain(m))
}
func deferableTestMain(m *testing.M) int {
	var (
		postgresUrl = flag.String("test_postgres_url", "", "postgres database url")
	)
	flag.Parse()
	timeout := 60 * time.Second
	testHasher = testutils.CreateHasherForTesting("testsalt")
	testRepo = &postgreSQLRepository{
		hasher:  testHasher,
		timeout: timeout,
	}
	if *testPostgreSQL {

		conn, err := createConnectionPool(*postgresUrl, timeout)
		if err != nil {
			log.Fatal(err)
			return 1
		}
		defer conn.Close()
		testRepo.conn = conn
		if err = testutils.TruncateAllTables(conn); err != nil {
			log.Fatal(err)
			return 1
		}

	}
	return m.Run()
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
	if *testPostgreSQL == false {
		return
	}
	hash, _ := testHasher.EncodeInt64([]int64{99})
	if _, err := testRepo.Find(hash); err != domain.ErrorURLNotFound {
		t.Fatal("Repo should return an URL not found error but got:", err)
	}
}

func TestFindReturnsUrl(t *testing.T) {
	if *testPostgreSQL == false {
		return
	}
	defer func(repo *postgreSQLRepository) {
		if err := testutils.TruncateUrlsTable(repo.conn); err != nil {
			t.Fatal("Failed to clean up after test:", err)
		}
	}(testRepo)

	expectedURL := "www.example.com"
	id, err := testutils.InsertUrl(testRepo.conn, expectedURL)
	if err != nil {
		t.Fatal("Failed to seed database", err)
	}
	hash, err := testHasher.EncodeInt64([]int64{id})
	if err != nil {
		t.Fatal("Unable to encode id")
	}

	res, err := testRepo.Find(hash)
	if err != nil {
		t.Fatal("Repo didn't find url:", err)
	}
	if res.Full != expectedURL {
		t.Fatal("Repo didn't find the correct url", res.Full, expectedURL)
	}

}

func TestCreateShouldReturnInvalidUrlError(t *testing.T) {
	if _, err := testRepo.Create(`javascript:alert("Hello World")`); err != domain.ErrorInvalidURL {
		t.Fatal("Repo should return an URL invalid error but got:", err)
	}
	if _, err := testRepo.Create(" "); err != domain.ErrorInvalidURL {
		t.Fatal("Repo should return an URL invalid error but got:", err)
	}
	if _, err := testRepo.Create(""); err != domain.ErrorInvalidURL {
		t.Fatal("Repo should return an URL invalid error but got:", err)
	}
}

func TestCreateShouldInsertInDatabase(t *testing.T) {
	if *testPostgreSQL == false {
		return
	}
	defer func(repo *postgreSQLRepository) {
		if err := testutils.TruncateUrlsTable(repo.conn); err != nil {
			t.Fatal("Failed to clean up after test:", err)
		}
	}(testRepo)

	url, err := testRepo.Create("www.example.com")
	if err != nil {
		t.Fatal("Repo shouldn't fail to create url:", err)
	}
	if url.Full != "http://www.example.com" {
		t.Fatal("http:// should be prepended", url.Full)
	}
	id, err := testutils.GeUrlId(testRepo.conn, url.Full, url.Hash)
	if err != nil {
		t.Fatal("Couldn't find url given return by repo in the database:", err)
	}
	ids, err := testHasher.DecodeInt64WithError(url.Hash)
	if err != nil {
		t.Fatal("Couldn't decode hash returned by repo", err)
	}
	if id != ids[0] {
		t.Fatal("Returned hash doesn't match with Id in database", id, ids[0])
	}
}

func TestCreateURLViewShouldReturnInvalidUrl(t *testing.T) {
	if err := testRepo.CreateURLView(""); err != domain.ErrorInvalidURL {
		t.Fatal("Repo should return an URL invalid error but got:", err)
	}
	if err := testRepo.CreateURLView(" "); err != domain.ErrorInvalidURL {
		t.Fatal("Repo should return an URL invalid error but got:", err)
	}
	if err := testRepo.CreateURLView("1"); err != domain.ErrorInvalidURL {
		t.Fatal("Repo should return an URL invalid error but got:", err)
	}
}

func TestCreateURLViewShouldInsertInDatabase(t *testing.T) {
	if *testPostgreSQL == false {
		return
	}
	defer func(repo *postgreSQLRepository) {
		if err := testutils.TruncateUrlViewsTable(repo.conn); err != nil {
			t.Fatal("Failed to clean up after test:", err)
		}
	}(testRepo)

	var id int64 = 4
	hash, err := testHasher.EncodeInt64([]int64{id})
	if err != nil {
		t.Fatal("Failed to hash id:", err)
	}

	if err := testRepo.CreateURLView(hash); err != nil {
		t.Fatal("Failed to add to database", err)
	}
	now := time.Now()
	view, err := testutils.GetUrlView(testRepo.conn, id)
	if err != nil {
		t.Fatal("Failed to get url view", err)
	}
	if now.Sub(view.CreatedAt) > 1*time.Hour {
		t.Fatal("CreatedAt timestamp not set correctly", view.CreatedAt, now)
	}

}

func TestStatsReturnsInvalidUrl(t *testing.T) {
	if _, err := testRepo.Stats(""); err != domain.ErrorInvalidURL {
		t.Fatal("Repo should return an URL invalid error but got:", err)
	}
	if _, err := testRepo.Stats(" "); err != domain.ErrorInvalidURL {
		t.Fatal("Repo should return an URL invalid error but got:", err)
	}
	if _, err := testRepo.Stats("1"); err != domain.ErrorInvalidURL {
		t.Fatal("Repo should return an URL invalid error but got:", err)
	}
}

func TestStatsCountsCorrectly(t *testing.T) {
	if *testPostgreSQL == false {
		return
	}
	defer func(repo *postgreSQLRepository) {
		if err := testutils.TruncateUrlViewsTable(repo.conn); err != nil {
			t.Fatal("Failed to clean up after test:", err)
		}
	}(testRepo)

	var id int64 = 8
	hash, err := testHasher.EncodeInt64([]int64{id})
	if err != nil {
		t.Fatal("Failed to encode", err)
	}

	// count 0, pastweek 0, pastday 0
	stats, err := testRepo.Stats(hash)
	if err != nil {
		t.Fatal("Failed to get stats", err)
	}
	if stats.Count != 0 || stats.PastWeekCount != 0 || stats.PastDayCount != 0 {
		t.Fatal("Stats count is not correct", stats)
	}

	// count 1, pastweek 0, pastday 0
	if err = testutils.InsertUrlView(testRepo.conn, id, time.Now().Add(-30*24*time.Hour)); err != nil {
		t.Fatal("Failed to insert with testutils", err)
	}
	stats, err = testRepo.Stats(hash)
	if err != nil {
		t.Fatal("Failed to get stats", err)
	}
	if stats.Count != 1 || stats.PastWeekCount != 0 || stats.PastDayCount != 0 {
		t.Fatal("Stats count is not correct", stats)
	}

	// count 2, pastweek 1, pastday 0
	if err = testutils.InsertUrlView(testRepo.conn, id, time.Now().AddDate(0, 0, -7)); err != nil {
		t.Fatal("Failed to insert with testutils", err)
	}
	stats, err = testRepo.Stats(hash)
	if err != nil {
		t.Fatal("Failed to get stats", err)
	}
	if stats.Count != 2 || stats.PastWeekCount != 1 || stats.PastDayCount != 0 {
		t.Fatal("Stats count is not correct", stats)
	}

	// count 3, pastweek 2, pastday 1
	if err = testutils.InsertUrlView(testRepo.conn, id, time.Now().AddDate(0, 0, -1)); err != nil {
		t.Fatal("Failed to insert with testutils", err)
	}
	stats, err = testRepo.Stats(hash)
	if err != nil {
		t.Fatal("Failed to get stats", err)
	}
	if stats.Count != 3 || stats.PastWeekCount != 2 || stats.PastDayCount != 1 {
		t.Fatal("Stats count is not correct", stats)
	}

}
