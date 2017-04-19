package core

import (
	"reflect"
	"testing"

	"github.com/OSSystems/hulk/unit"
	"github.com/bouk/monkey"
	"github.com/stretchr/testify/assert"
)

func TestNewHooker(t *testing.T) {
	s := NewSubscriber()

	assert.NotNil(t, s.Hooks)
}

func TestHookerHooks(t *testing.T) {
	s := NewSubscriber()

	s.unit = &unit.Subscriber{
		Topics: []string{},
		Hooks: unit.Hooks{
			OnPublish:       "OnPublish",
			OnSubscribe:     "OnSubscribe",
			OnSubscribeFail: "OnSubscribeFail",
		},
	}

	callStack := []string{}

	defer monkey.PatchInstanceMethod(reflect.TypeOf(s), "ExecuteHook", func(_ *Subscriber, cmdLine string, payload []byte) error {
		callStack = append(callStack, cmdLine)
		return nil
	}).Unpatch()

	s.Hooks.OnPublish("", []byte(""))
	s.Hooks.OnSubscribe("")
	s.Hooks.OnSubscribeFail("")

	expectedCallStack := []string{"OnPublish", "OnSubscribe", "OnSubscribeFail"}

	assert.Equal(t, expectedCallStack, callStack)
}
