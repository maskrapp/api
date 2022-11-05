package utils_test

import (
	"testing"

	"github.com/maskrapp/backend/utils"
	"github.com/stretchr/testify/assert"
)

func TestPasswordValidation(t *testing.T) {
	assert.Equal(t, utils.IsValidPassword("weakpassword"), false)
	assert.Equal(t, utils.IsValidPassword("weak_password"), false)
	assert.Equal(t, utils.IsValidPassword("weak_password123"), false)
	assert.Equal(t, utils.IsValidPassword("Stronger_password123"), true)
	assert.Equal(t, utils.IsValidPassword("Extremely_Long_Password123456789!"), false)
	assert.Equal(t, utils.IsValidPassword("Stronger Password_123"), false)
}
