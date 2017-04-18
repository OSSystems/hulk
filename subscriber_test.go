package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubscriberExpandTopics(t *testing.T) {
	s := NewSubscriber()

	s.Environment = map[string]string{
		"VARIABLE": "value",
	}

	s.Topics = []string{"{VARIABLE}"}

	err := s.ExpandTopics()
	assert.NoError(t, err)

	assert.Equal(t, []string{s.Environment["VARIABLE"]}, s.Topics)
}
