package fng

import "net/http"

// Fear and Greed Index API
// https://alternative.me/crypto/fear-and-greed-index/
// https://api.alternative.me/fng/?limit=100&format=json&date_format=cn
type FearAndGreedIndex struct {
	client *http.Client
}
