package vlc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListDevices(t *testing.T) {
	Initialize()
	_, err := GetDevices()
	assert.NoError(t, err)
}
