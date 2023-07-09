package storage

import "errors"

var KeyError = errors.New("key error")

type Interface interface {
	Load(k string) ([]byte, error)
	Save(k string, t []byte) error
	Delete(k string) error
}
