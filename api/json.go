package api

import (
	domain "github.com/yanisky/url-shortener/pkg"
)

type urlStats struct {
	URL   domain.URL          `json:"url"`
	Views domain.URLViewStats `json:"views"`
}

type urlCreatedJsonResponse struct {
	Data domain.URL `json:"data"`
}

type urlStatsJsonResponse struct {
	Data urlStats `json:"data"`
}

type errorJsonResponse struct {
	Message string `json:"message"`
}
