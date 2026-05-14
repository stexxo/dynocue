// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateModel(t *testing.T) {
	t.Parallel()

	m, err := NewCueingModel()
	assert.NoError(t, err)
	assert.NotNil(t, m)
}
