package hulk

import (
	"fmt"
	"os/exec"

	"github.com/OSSystems/pkg/log"
	"github.com/Sirupsen/logrus"
)

// HookName holds the supported hooks
type HookName int

const (
	// OnReceiveHook represents the OnReceive hook
	OnReceiveHook = iota
)

var hookNames = map[HookName]string{
	OnReceiveHook: "OnReceiveHook",
}

// Hook is the hook representation
type Hook struct {
	service *Service
	name    HookName
	topic   string
}

// NewHook creates a new Hook instance
func NewHook(service *Service, name HookName, topic string) *Hook {
	hook := &Hook{
		service: service,
		name:    name,
		topic:   topic,
	}

	if hook.cmdLine() == "" {
		return nil
	}

	return hook
}

// cmdLine returns the cmd line of the hook
func (h *Hook) cmdLine() string {
	switch h.name {
	case OnReceiveHook:
		return h.service.manifest.Hooks.OnReceive
	}

	return ""
}

// createCmd creates command
func (h *Hook) createCmd() *exec.Cmd {
	args := []string{"-c"}
	args = append(args, h.cmdLine())

	cmd := exec.Command("sh", args...)
	cmd.Env = []string{}

	for key, value := range h.service.environment {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	cmd.Env = append(cmd.Env, fmt.Sprintf("TOPIC=%s", h.topic))

	return cmd
}

// execute executes hook command
func (h *Hook) execute(payload []byte) error {
	cmd := h.createCmd()

	logFields := logrus.Fields{
		"service": h.service.name,
		"hook":    HookNameToString(h.name),
	}

	if log.GetLevel() == logrus.DebugLevel {
		logFields["cmd"] = h.cmdLine()
		logFields["env"] = cmd.Env
		logFields["payload"] = string(payload)
		log.WithFields(logFields).Debug("executing hook")
	} else {
		log.WithFields(logFields).Info("executing hook")
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	defer stdin.Close()

	err = cmd.Start()
	if err != nil {
		return err
	}

	stdin.Write(payload)

	return nil
}

// HookNameToString converts hook name to string
func HookNameToString(name HookName) string {
	return hookNames[name]
}
