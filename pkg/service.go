package urlshortener

type URLShortenerService interface {
	Find(hashUrl string, shouldTrack bool) (URL, error)
	Create(url string) (URL, error)
	RecordURLView(urlHash string) error
	Stats(hashUrl string) (URLViewStats, error)
}

type urlShortenerService struct {
	store     URLStoreRepository
	analytics URLAnalyticsRepository
}

// Find will find the url that matches the short url hash
// if shouldTrack is set to true it add a view to the url
// to get the stats use the Stats method
func (s *urlShortenerService) Find(urlHash string, shouldTrack bool) (URL, error) {
	url, err := s.store.Find(urlHash)
	if err != nil {
		return URL{}, err
	}
	if shouldTrack == true {
		go func() {
			s.RecordURLView(urlHash)
		}()
	}
	return url, nil
}

// Create creates a short url hash that can be used in the service
func (s *urlShortenerService) Create(fullUrl string) (URL, error) {
	url, err := s.store.Create(fullUrl)
	if err != nil {
		return URL{}, err
	}

	return url, nil
}

func (s *urlShortenerService) RecordURLView(urlHash string) error {
	return s.analytics.CreateURLView(urlHash)
}

// Stats returns some basic stats
// count is the total times the url has been used
// pastWeekCount is the total times the url has been used in the past week
// pastDayCount is the total times the url has been used in the past 24h
func (s *urlShortenerService) Stats(urlHash string) (URLViewStats, error) {
	return s.analytics.Stats(urlHash)
}

func NewURLShortenerService(store URLStoreRepository, analytics URLAnalyticsRepository) URLShortenerService {
	return &urlShortenerService{
		store:     store,
		analytics: analytics,
	}
}
