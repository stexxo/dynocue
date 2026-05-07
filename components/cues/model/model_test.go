package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateModel(t *testing.T) {
	m, err := NewCueingModel()
	assert.NoError(t, err)
	assert.NotNil(t, m)
}
