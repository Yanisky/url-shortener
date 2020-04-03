package urlshortener

type cachedURLShortenerService struct {
	cache   URLCacheRepository
	service URLShortenerService
}

// Find will try to find url from cache first than from service
// if a url was found it will cache it automatically
func (s *cachedURLShortenerService) Find(urlHash string, shouldTrack bool) (URL, error) {
	url, err := s.cache.Find(urlHash)
	if err != nil { // cache miss
		url, err = s.service.Find(urlHash, shouldTrack) // get from url service
		if err != nil {
			return URL{}, err
		}
		// save in cache
		go func(toCache URL) {
			s.cache.Cache(url)
		}(url)
		return url, nil
	}

	if shouldTrack == true {
		go func(hash string) {
			s.service.RecordURLView(hash)
		}(urlHash)
	}

	return url, nil
}

// Create creates a short url and caches the value
func (s *cachedURLShortenerService) Create(fullUrl string) (URL, error) {
	url, err := s.service.Create(fullUrl)
	if err != nil {
		return URL{}, err
	}

	go func() {
		s.cache.Cache(url)
	}()

	return url, nil
}
func (s *cachedURLShortenerService) RecordURLView(urlHash string) error {
	return s.service.RecordURLView(urlHash)
}

func (s *cachedURLShortenerService) Stats(urlHash string) (URLViewStats, error) {
	return s.service.Stats(urlHash)
}

func NewCachedURLShortenerService(service URLShortenerService, cacheRepo URLCacheRepository) URLShortenerService {
	return &cachedURLShortenerService{
		service: service,
		cache:   cacheRepo,
	}
}
