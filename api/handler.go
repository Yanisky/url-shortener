package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	domain "github.com/yanisky/url-shortener/pkg"
)

type URLShortnerHttpHandler interface {
	Redirect(http.ResponseWriter, *http.Request)
	CreateURL(http.ResponseWriter, *http.Request)
	ViewUrlStats(http.ResponseWriter, *http.Request)
}

type handler struct {
	urlService domain.URLShortenerService
}

// NewGorillaHTTPHandler creates a new http handler that works with Gorilla's router
func NewGorillaHTTPHandler(service domain.URLShortenerService) URLShortnerHttpHandler {
	return &handler{urlService: service}
}

// Redirect URL hash to its full URL
func (h *handler) Redirect(response http.ResponseWriter, request *http.Request) {
	urlHash, ok := mux.Vars(request)["urlHash"]
	if ok == false {
		chooseErrorResponse(domain.ErrorURLNotFound, response)
		return
	}
	url, err := h.urlService.Find(urlHash, true)
	// The switch is here to return 404 when the url has an invalid hash, that's not information the user needs to know.
	switch err {
	case nil:
		http.Redirect(response, request, url.Full, http.StatusMovedPermanently)
	case domain.ErrorInvalidURL:
		chooseErrorResponse(domain.ErrorURLNotFound, response)
	default:
		chooseErrorResponse(err, response)
	}
}

// Create a new URL
func (h *handler) CreateURL(response http.ResponseWriter, request *http.Request) {
	type createShortURLRequest struct {
		URL string `json:"url"`
	}
	data := &createShortURLRequest{}

	if err := json.NewDecoder(request.Body).Decode(data); err != nil {
		chooseErrorResponse(err, response)
		return
	}

	url, err := h.urlService.Create(data.URL)
	if err != nil {
		chooseErrorResponse(err, response)
		return
	}

	responseData, err := json.Marshal(&urlCreatedJsonResponse{
		Data: url,
	})
	if err != nil {
		chooseErrorResponse(err, response)
		return
	}
	response.WriteHeader(http.StatusOK)
	response.Write(responseData)
}

// ViewUrlStats returns stats for the given url hash
func (h *handler) ViewUrlStats(response http.ResponseWriter, request *http.Request) {
	urlHash, ok := mux.Vars(request)["urlHash"]
	if ok == false {
		chooseErrorResponse(domain.ErrorURLNotFound, response)
		return
	}
	url, err := h.urlService.Find(urlHash, false)
	if err != nil {
		chooseErrorResponse(err, response)
		return
	}

	stats, err := h.urlService.Stats(urlHash)
	if err != nil {
		chooseErrorResponse(err, response)
		return
	}

	responseData, err := json.Marshal(&urlStatsJsonResponse{
		Data: urlStats{
			URL:   url,
			Views: stats,
		},
	})
	if err != nil {
		chooseErrorResponse(err, response)
		return
	}
	response.WriteHeader(http.StatusOK)
	response.Write(responseData)
}

func chooseErrorResponse(err error, response http.ResponseWriter) {
	response.Header().Set("Content-Type", "application/json; charset=utf-8")
	switch err {
	case domain.ErrorInvalidURL:
		response.WriteHeader(http.StatusBadRequest)
	case domain.ErrorURLNotFound:
		response.WriteHeader(http.StatusNotFound)
	default:
		response.WriteHeader(http.StatusInternalServerError)
		err = errors.New("Internal Server Error")
	}
	json.NewEncoder(response).Encode(errorJsonResponse{Message: err.Error()})
}
