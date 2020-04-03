package urlshortener

import (
	"errors"
	"testing"
	"time"
)

func TestFindReturnsCorrectly(t *testing.T) {
	t.Parallel()
	mockUrl := &URL{Full: "Full"}
	expectedHash := "HASH"
	repoMock := &urlShortenerRepoMock{url: mockUrl}
	service := NewURLShortenerService(repoMock, repoMock)

	url, err := service.Find(expectedHash, false)
	if err != nil {
		t.Fatal("Service should not fail to find the hash")
	}
	if url.Full != mockUrl.Full || url.Hash != expectedHash {
		t.Fatal("Service should not fail to find the hash")
	}
	time.Sleep(100 * time.Millisecond)
	if repoMock.createURLViewCalled == true {
		t.Fatal("Service should NOT have called createURLView function")
	}
}

func TestFindTracksViews(t *testing.T) {
	t.Parallel()
	repoMock := &urlShortenerRepoMock{url: &URL{}}
	expectedHash := "HASH"
	service := NewURLShortenerService(repoMock, repoMock)
	url, err := service.Find(expectedHash, true)
	if err != nil {
		t.Fatal("Service should not fail to find the hash")
	}
	time.Sleep(100 * time.Millisecond)
	if url.Hash != expectedHash {
		t.Fatal("'Find didn't return the correct URL", url)
	}
	if repoMock.createURLViewCalled == false {
		t.Fatal("Service should have called createURLView function")
	}

}

func TestFindBubblesUpError(t *testing.T) {
	expectedError := errors.New("Bubble up")
	repoMock := &urlShortenerRepoMock{err: expectedError}
	service := NewURLShortenerService(repoMock, repoMock)
	_, err := service.Find("hash", false)
	if err != expectedError {
		t.Fatal("Service should have failed with the expected error", err)
	}
}

func TestCreate(t *testing.T) {
	expectedHash := "CreateHash"
	expectedUrl := "CreateURL"
	repoMock := &urlShortenerRepoMock{url: &URL{Hash: expectedHash}}
	service := NewURLShortenerService(repoMock, repoMock)
	url, err := service.Create(expectedUrl)
	if err != nil {
		t.Fatal("Service shouldn't have failed:", err)
	}
	if url.Full != expectedUrl || url.Hash != expectedHash {
		t.Fatal("Service is not using the repository", url)
	}
}

func TestCreateBubblesError(t *testing.T) {
	expectedError := errors.New("Bubble up")
	repoMock := &urlShortenerRepoMock{err: expectedError}
	service := NewURLShortenerService(repoMock, repoMock)
	_, err := service.Create("hash")
	if err != expectedError {
		t.Fatal("Service should have failed with the expected error", err)
	}
}

func TestStats(t *testing.T) {
	expectedStats := &URLViewStats{
		Count:         99,
		PastWeekCount: 98,
		PastDayCount:  97,
	}
	expectedHash := "expected-hash"
	url := &URL{}
	repoMock := &urlShortenerRepoMock{stats: expectedStats, url: url}
	service := NewURLShortenerService(repoMock, repoMock)
	stats, err := service.Stats(expectedHash)
	if err != nil {
		t.Fatal("Service should have not fail on stats", err)
	}
	if stats.Count != expectedStats.Count || stats.PastWeekCount != expectedStats.PastWeekCount || stats.PastDayCount != expectedStats.PastDayCount {
		t.Fatal("Service returned wrong stats", stats, expectedStats)
	}

}

func TestStatsBubblesUpError(t *testing.T) {
	expectedError := errors.New("Bubble up")
	repoMock := &urlShortenerRepoMock{err: expectedError}
	service := NewURLShortenerService(repoMock, repoMock)
	_, err := service.Stats("hash")
	if err != expectedError {
		t.Fatal("Service should have failed with the expected error", err)
	}
}

type urlShortenerRepoMock struct {
	url                 *URL
	stats               *URLViewStats
	err                 error
	createURLViewCalled bool
	memorizeCalled      bool
}

func (r *urlShortenerRepoMock) Find(urlHash string) (URL, error) {
	if r.url != nil {
		r.url.Hash = urlHash
		return *r.url, nil
	}
	return URL{}, r.err
}
func (r *urlShortenerRepoMock) Create(url string) (URL, error) {
	if r.url != nil {
		r.url.Full = url
		return *r.url, nil
	}
	return URL{}, r.err
}
func (r *urlShortenerRepoMock) CreateURLView(urlHash string) error {
	r.createURLViewCalled = true
	if r.url != nil {
		r.url.Hash = urlHash
		return nil
	}
	return r.err
}
func (r *urlShortenerRepoMock) Stats(urlHash string) (URLViewStats, error) {
	if r.stats != nil {
		r.url = &URL{Hash: urlHash}
		return *r.stats, nil
	}
	return URLViewStats{}, r.err
}
