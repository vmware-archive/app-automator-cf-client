package models

import "net/http"

type HeaderOption func(header *http.Header)
