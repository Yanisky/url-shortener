package urlshortener

import (
	"errors"
)

var (
	ErrorURLNotFound = errors.New("URL Not Found")
	ErrorInvalidURL  = errors.New("Invalid URL")
)
