package main

import (
	"crypto/rand"

	"github.com/dgraph-io/badger"
)

type CodeStorage struct {
	*badger.DB
}

func Open(folder string) (*CodeStorage, error) {
	opts := badger.DefaultOptions
	opts.Dir = folder
	opts.Truncate = true
	opts.ValueDir = folder
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	return &CodeStorage{db}, nil
}

func (s *CodeStorage) GetCode(id string) (string, error) {
	txn := s.NewTransaction(false)
	defer txn.Commit(nil)
	item, err := txn.Get([]byte(id))
	if err != nil {
		return "", err
	}
	code, err := item.ValueCopy(nil)
	return string(code), err
}

func (s *CodeStorage) SaveCode(code []byte) (string, error) {
	txn := s.NewTransaction(true)
	defer txn.Commit(nil)

	// generate random id
	var id []byte
	var err error
	for err != badger.ErrKeyNotFound {
		id = GetRandomBytes(11)
		_, err = txn.Get(id)
		if err != nil && err != badger.ErrKeyNotFound {
			return "", err
		}
	}

	if err := txn.Set(id, code); err != nil {
		return "", err
	}
	return string(id), nil
}

// GetRandomBytes generates base62 string of given length
// Available characters are: "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
func GetRandomBytes(n int) []byte {
	const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, n)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return bytes
}