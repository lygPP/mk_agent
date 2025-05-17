package httputil

import (
	"net/http"
)

type Config struct {
	URL        string
	APIKey     string
	HTTPClient *http.Client
}
