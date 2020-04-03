package urlshortener

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func TestFindUsesCache(t *testing.T) {
	service := &urlshortenerServiceMock{}
	url := URL{Full: "Full URL"}
	cacheRepo := &urlCacheRepoMock{url: url}
	cachedService := NewCachedURLShortenerService(service, cacheRepo)
	expectedHash := "hash"
	cache, err := cachedService.Find(expectedHash, false)
	if err != nil {
		t.Fatal("Failed to get from cache:", err)
	}
	if cacheRepo.findCalled == false || cacheRepo.hash != expectedHash || cache.Full != url.Full {
		t.Fatal("Cache repo was not used correctly", cacheRepo)
	}
	if service.findCalled == true {
		t.Fatal("Service was used when cache was available, it will add server load and degrade performance")
	}
}
func TestFindFallsbackToServiceWhenCacheFails(t *testing.T) {
	t.Parallel()
	service := &urlshortenerServiceMock{url: URL{Full: "Full URL"}}
	cacheRepo := &urlCacheRepoMock{err: errors.New("cache error")}
	cachedService := NewCachedURLShortenerService(service, cacheRepo)
	expectedHash := "hash"
	cache, err := cachedService.Find(expectedHash, true)
	if err != nil {
		t.Fatal("Failed to get from cache service:", err)
	}
	if service.findCalled == false || service.val != expectedHash || service.url.Full != cache.Full || service.shouldTrack == false {
		t.Fatal("Service was used when cache was available, it will add server load and degrade performance")
	}
	time.Sleep(100 * time.Millisecond)
	// check that service caches fallback result
	if cacheRepo.cacheCalled == false || cacheRepo.url.Full != service.url.Full {
		t.Fatal("Service should have cached result automatically", cacheRepo)
	}
}

func TestCachedFindRecordsUrlsViewsOnce(t *testing.T) {
	t.Parallel()
	service := &urlshortenerServiceMock{}
	cacheRepo := &urlCacheRepoMock{}
	cachedService := NewCachedURLShortenerService(service, cacheRepo)
	expectedHash := "hash"
	_, err := cachedService.Find(expectedHash, true)
	if err != nil {
		t.Fatal("Failed to get from cache service:", err)
	}
	time.Sleep(100 * time.Millisecond)
	if service.recordCalled > 1 || service.val != expectedHash || service.findCalled == true {
		t.Fatal("Cached service is not recording views correctly", service)
	}
	// when cache fails
	service = &urlshortenerServiceMock{}
	cacheRepo = &urlCacheRepoMock{err: errors.New("cache error")}
	cachedService = NewCachedURLShortenerService(service, cacheRepo)

	_, err = cachedService.Find(expectedHash, true)
	if err != nil {
		t.Fatal("Failed to get from cache service:", err)
	}
	time.Sleep(100 * time.Millisecond)
	if service.recordCalled > 1 || service.val != expectedHash || service.findCalled == false {
		t.Fatal("Cached service is not recording views correctly", service)
	}
}

func TestCreateCachesResult(t *testing.T) {
	t.Parallel()
	service := &urlshortenerServiceMock{url: URL{Hash: "hash"}}
	cacheRepo := &urlCacheRepoMock{}
	cachedService := NewCachedURLShortenerService(service, cacheRepo)
	expectedURL := "www.example.com"
	cache, err := cachedService.Create(expectedURL)
	if err != nil {
		t.Fatal("Failed to create from cached service:", err)
	}
	time.Sleep(100 * time.Millisecond)
	if service.createCalled == false || service.val != expectedURL || cacheRepo.cacheCalled == false || cache.Hash != service.url.Hash {
		t.Fatal("Failed to create and cache:", service, cacheRepo)
	}
}
func TestCreateDoesntCacheIfError(t *testing.T) {
	t.Parallel()
	service := &urlshortenerServiceMock{err: errors.New("service error")}
	cacheRepo := &urlCacheRepoMock{}
	cachedService := NewCachedURLShortenerService(service, cacheRepo)
	expectedURL := "www.example.com"
	_, err := cachedService.Create(expectedURL)
	if err != service.err {
		t.Fatal("Cached service should have returned an error", err)
	}
	time.Sleep(100 * time.Millisecond)
	if service.createCalled == false || cacheRepo.cacheCalled == true {
		t.Fatal("Cached was called on error", service, cacheRepo)
	}
}

type urlshortenerServiceMock struct {
	m            sync.Mutex
	findCalled   bool
	createCalled bool
	recordCalled int
	statsCalled  bool
	shouldTrack  bool
	val          string
	err          error
	stats        URLViewStats
	url          URL
}

func (s *urlshortenerServiceMock) Find(urlHash string, shouldTrack bool) (URL, error) {
	s.findCalled = true
	s.shouldTrack = shouldTrack
	s.val = urlHash

	return s.url, s.err
}
func (s *urlshortenerServiceMock) Create(url string) (URL, error) {
	s.createCalled = true
	s.val = url
	return s.url, s.err
}
func (s *urlshortenerServiceMock) RecordURLView(urlHash string) error {
	s.m.Lock()
	s.recordCalled = s.recordCalled + 1
	s.val = urlHash
	s.m.Unlock()
	return s.err
}
func (s *urlshortenerServiceMock) Stats(urlHash string) (URLViewStats, error) {
	s.findCalled = true
	s.val = urlHash
	return s.stats, s.err
}

type urlCacheRepoMock struct {
	findCalled  bool
	cacheCalled bool
	url         URL
	hash        string
	err         error
}

func (r *urlCacheRepoMock) Find(urlHash string) (URL, error) {
	r.findCalled = true
	r.hash = urlHash
	return r.url, r.err
}
func (r *urlCacheRepoMock) Cache(url URL) error {
	r.cacheCalled = true
	r.url = url
	return r.err
}
