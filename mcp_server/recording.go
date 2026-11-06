package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type recordActionInput struct {
	Serial         string `json:"serial" jsonschema:"device serial"`
	Name           string `json:"name" jsonschema:"<app>_<screen>_<what>[_variant]"`
	Timeout        int    `json:"timeout,omitempty" jsonschema:"ms without a touch or key press that ends the recording, default 5000"`
	DeviceSpecific bool   `json:"device_specific,omitempty" jsonschema:"true only when the user wants the action for this device alone; the server appends _<model_os> to name"`
}

func (s *Server) handleRecordAction(
	ctx context.Context,
	req *mcp.CallToolRequest,
	in recordActionInput,
) (*mcp.CallToolResult, any, error) {
	if in.Serial == "" || in.Name == "" {
		return nil, nil, fmt.Errorf("serial and name are required")
	}
	name, err := s.libraryName(in.Serial, in.Name, in.DeviceSpecific)
	if err != nil {
		return nil, nil, err
	}
	recording, err := s.recordings.start(in.Serial, name)
	if err != nil {
		return nil, nil, err
	}
	go s.runRecording(in.Serial, name, in.Timeout, recording.done)
	return s.finishRecording(ctx, in.Serial, recording, toolWaitSec*time.Second)
}

func (s *Server) finishRecording(
	ctx context.Context,
	serial string,
	recording pendingRecording,
	wait time.Duration,
) (*mcp.CallToolResult, any, error) {
	outcome, finished := awaitRecording(ctx, recording.done, wait)
	if !finished {
		return recordingPending(recording.name), nil, nil
	}
	s.recordings.finish(serial)
	return recordingResult(recording.name, outcome)
}

const nothingRecorded = "nothing recorded: no touch or hardware key happened on the device within " +
	"the timeout after the call; ask the user to perform the gesture again"

type recordOutcome struct {
	saved bool
	err   error
}

type pendingRecording struct {
	name string
	done chan recordOutcome
}

type recordings struct {
	mutex   sync.Mutex
	pending map[string]pendingRecording
}

func (r *recordings) start(serial string, name string) (pendingRecording, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	previous, ok := r.pending[serial]
	if ok && len(previous.done) == 0 {
		return pendingRecording{}, fmt.Errorf(
			"a recording is already running on %s; call wait_for_session for its result", serial,
		)
	}
	if r.pending == nil {
		r.pending = map[string]pendingRecording{}
	}
	recording := pendingRecording{name: name, done: make(chan recordOutcome, 1)}
	r.pending[serial] = recording
	return recording, nil
}

func (r *recordings) get(serial string) (pendingRecording, bool) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	recording, ok := r.pending[serial]
	return recording, ok
}

func (r *recordings) finish(serial string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	delete(r.pending, serial)
}

func (s *Server) runRecording(serial string, name string, timeout int, done chan<- recordOutcome) {
	saved, err := s.api.recordAction(serial, name, timeout)
	done <- recordOutcome{saved: saved, err: err}
}

func awaitRecording(
	ctx context.Context,
	done <-chan recordOutcome,
	timeout time.Duration,
) (recordOutcome, bool) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case outcome := <-done:
		return outcome, true
	case <-timer.C:
		return recordOutcome{}, false
	case <-ctx.Done():
		return recordOutcome{err: ctx.Err()}, true
	}
}

func recordingResult(name string, outcome recordOutcome) (*mcp.CallToolResult, any, error) {
	if outcome.err != nil {
		return nil, nil, outcome.err
	}
	if !outcome.saved {
		return textResult(nothingRecorded), nil, nil
	}
	return textResult("saved action " + name), nil, nil
}

func recordingPending(name string) *mcp.CallToolResult {
	return textResult(fmt.Sprintf(
		"recording %s still running after %ds - call wait_for_session for its result",
		name, toolWaitSec,
	))
}
