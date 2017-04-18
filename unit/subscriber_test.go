package unit

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubscriberFromData(t *testing.T) {
	unit, err := SubscriberFromData([]byte(""))
	assert.NoError(t, err)
	assert.NotNil(t, unit)

	unit, err = SubscriberFromData([]byte("invalid"))
	assert.Error(t, err)
	assert.Nil(t, unit)
}
