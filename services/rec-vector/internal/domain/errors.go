package domain

import "errors"

var (
	ErrNotFound            = errors.New("not found")
	ErrCollectionNotFound  = errors.New("collection not found")
	ErrModelNotFound       = errors.New("embedding model not found")
	ErrMetadataNotFound    = errors.New("vector metadata not found")
	ErrJobNotFound         = errors.New("index job not found")
	ErrUnauthorized        = errors.New("unauthorized")
)
