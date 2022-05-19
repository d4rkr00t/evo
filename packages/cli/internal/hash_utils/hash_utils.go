package hash_utils

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
)

func HashStringList(list []string) string {
	var h = sha1.New()

	for _, item := range list {
		io.WriteString(h, item)
	}

	return hex.EncodeToString(h.Sum(nil))
}
