package core

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"

	"github.com/bouk/monkey"
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

func TestSubscriberExecuteHook(t *testing.T) {
	s := NewSubscriber()

	var stdout io.ReadCloser
	var guard *monkey.PatchGuard

	guard = monkey.PatchInstanceMethod(reflect.TypeOf(s), "CreateHookCommand", func(_ *Subscriber, cmdLine string) *exec.Cmd {
		guard.Unpatch()
		defer guard.Restore()

		cmd := s.CreateHookCommand(cmdLine)
		stdout, _ = cmd.StdoutPipe()

		return cmd
	})

	expectedDump := processDump{
		Stdin: "payload",
		Args:  []string{"arg1", "arg2"},
		Env:   []string{"GO_WANT_HELPER_PROCESS=1"},
	}

	cmdLine := []string{os.Args[0], "-test.run=TestHelperProcess", "--"}
	cmdLine = append(cmdLine, expectedDump.Args...)

	for _, env := range expectedDump.Env {
		slices := strings.Split(env, "=")
		s.Environment[slices[0]] = slices[1]
	}

	err := s.ExecuteHook(strings.Join(cmdLine, " "), []byte(expectedDump.Stdin))
	assert.NoError(t, err)

	bytes, _ := ioutil.ReadAll(stdout)

	dump := processDump{}

	err = json.Unmarshal(bytes, &dump)
	assert.NoError(t, err)

	assert.Equal(t, expectedDump, dump)
}

func TestHelperProcess(*testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	defer os.Exit(0)

	args := os.Args
	for len(args) > 0 {
		if args[0] == "--" {
			args = args[1:]
			break
		}
		args = args[1:]
	}

	stdin, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		os.Exit(1)
	}

	dump := &processDump{
		Stdin: string(stdin),
		Args:  args,
		Env:   os.Environ(),
	}

	bytes, err := json.Marshal(dump)
	if err != nil {
		os.Exit(1)
	}

	fmt.Print(string(bytes))
}

type processDump struct {
	Stdin string
	Args  []string
	Env   []string
}
