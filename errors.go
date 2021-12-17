package octopusapi

import "errors"

var (
	// ErrURLParseFailed defines URL parse failed error.
	ErrURLParseFailed = errors.New("URL parse failed")
	// ErrProductDetailParseFailed defines product detail parse failed error.
	ErrProductDetailParseFailed = errors.New("product detail parse failed")
	// ErrProductDetailParseFailed defines image ID parse failed error.
	ErrImageIDParseFailed = errors.New("image ID parse failed")
)
