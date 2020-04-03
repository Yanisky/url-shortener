package testutils

import "github.com/speps/go-hashids"

func CreateHasherForTesting(salt string) *hashids.HashID {
	hd := hashids.NewData()
	hd.Salt = salt
	hd.MinLength = 7
	h, _ := hashids.NewWithData(hd)
	return h
}
