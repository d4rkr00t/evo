package label_test

import (
	"evo/internal/label"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Get_Labels_From_Full_Specifier(t *testing.T) {
	var lls, _ = label.GetLablesFromList([]string{"workspace::build"}, "*")
	assert.Equal(t, lls[0].Scope, "workspace")
	assert.Equal(t, lls[0].Target, "build")
}

func Test_Get_Labels_From_Target_Specifier(t *testing.T) {
	var lls, _ = label.GetLablesFromList([]string{"::build"}, "*")
	assert.Equal(t, lls[0].Scope, "*")
	assert.Equal(t, lls[0].Target, "build")
}

func Test_Get_Labels_Errors_With_Only_Scope(t *testing.T) {
	var _, err = label.GetLablesFromList([]string{"workspace"}, "*")
	assert.Error(t, err)
}
