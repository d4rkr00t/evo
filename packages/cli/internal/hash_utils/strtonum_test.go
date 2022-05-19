package hash_utils_test

import (
	"evo/internal/hash_utils"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_StrToNum(t *testing.T) {
	assert.Equal(t, hash_utils.StrToNum("hello"), 760)
}
