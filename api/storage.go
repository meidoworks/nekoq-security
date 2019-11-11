package api

import "errors"

var ErrNotFound = errors.New("key not found")

type Storage interface {
	GetObject(key string) (interface{}, error)
	PutObject(key string, obj interface{}) error
}
