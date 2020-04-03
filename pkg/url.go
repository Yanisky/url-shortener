package urlshortener

import (
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
)

// URL struct
type URL struct {
	Hash      string    `json:"hash"`
	Full      string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`
}

// URLViewStats struct
type URLViewStats struct {
	PastDayCount  int `json:"past_day_count"`
	PastWeekCount int `json:"past_week_count"`
	Count         int `json:"count"`
}

// NormalizeURL checks if the given string is a valid url
// and prepends http:// if it doesn't start with 'http://' or 'https://'
func NormalizeURL(url string) (string, error) {
	if ok := govalidator.MinStringLength(url, "1"); ok == false {
		return "", ErrorInvalidURL
	}
	// Internet explorer has a limit of 2048 characters in the address bar
	// this might no longer be the case but we should have a limit and 2048 seems reasonable.
	if ok := govalidator.MaxStringLength(url, "2048"); ok == false {
		return "", ErrorInvalidURL
	}

	loweredURL := strings.ToLower(url)
	if !strings.HasPrefix(loweredURL, "http://") && !strings.HasPrefix(loweredURL, "https://") {
		url = "http://" + url
	}

	// This is a very loose URL validation. It should be replaced by a custom one.
	if ok := govalidator.IsRequestURL(url); ok == false {
		return "", ErrorInvalidURL
	}

	return url, nil
}
