package urlshortener

type URLCacheRepository interface {
	Find(urlHash string) (URL, error)
	Cache(url URL) error
}

type URLStoreRepository interface {
	Find(urlHash string) (URL, error)
	Create(url string) (URL, error)
}

type URLAnalyticsRepository interface {
	CreateURLView(urlHash string) error
	Stats(urlHash string) (URLViewStats, error)
}
