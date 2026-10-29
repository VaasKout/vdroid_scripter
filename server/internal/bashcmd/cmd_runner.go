// Package bashcmd ...
package bashcmd

import (
	"android_vision_scripter/internal/filesdb"
	"android_vision_scripter/pkg/logger"
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// CmdAPI ...
type CmdAPI interface {
	AdbAPI
	ScrCpy
}
type cmdImpl struct {
	logger  *logger.Logger
	filesDB filesdb.FilesDB

	scrcpyMu    sync.Mutex
	scrcpyProcs map[string]*backgroundProc
}

type backgroundProc struct {
	cmd  *exec.Cmd
	done chan struct{}
}

// New instance of CmdAPI
func New(
	filesDB filesdb.FilesDB,
	logger *logger.Logger,
) CmdAPI {
	return &cmdImpl{
		logger:      logger,
		filesDB:     filesDB,
		scrcpyProcs: make(map[string]*backgroundProc),
	}
}

func (c *cmdImpl) ExecuteCommand(cmd string) (string, error) {
	if cmd == "" {
		return "", fmt.Errorf("cmd is empty")
	}
	cmdExec := exec.Command("bash", "-c", cmd)
	cmdExec.Stdin = os.Stdin
	cmdExec.Stderr = os.Stderr
	c.logger.Info("-------")
	c.logger.Info(fmt.Sprintf("(%s): Start... ⏳", cmd))
	result, err := cmdExec.Output()
	if len(result) > 0 {
		var trimmedResult = strings.Trim(string(result), "\n")
		c.logger.Info(fmt.Sprintf("(%s): %s ✅", cmd, trimmedResult))
	} else {
		c.logger.Info(fmt.Sprintf("(%s): DONE ✅", cmd))
	}

	if err != nil {
		c.logger.Error(fmt.Sprintf("(%s): %s ❌", cmd, err.Error()))
	}
	return string(result), err
}

func (c *cmdImpl) executeUntilIdle(
	cmd string,
	idle time.Duration,
	limit time.Duration,
	axesByDevice map[string]touchAxes,
) (string, error) {
	if cmd == "" {
		return "", fmt.Errorf("cmd is empty")
	}
	ctx, cancel := context.WithTimeout(context.Background(), limit)
	defer cancel()

	c.logger.Info("-------")
	c.logger.Info(fmt.Sprintf("(%s): Start until %s without input... ⏳", cmd, idle))
	cmdExec, lines, err := startLineStream(ctx, cmd)
	if err != nil {
		c.logger.Error(fmt.Sprintf("(%s): %s ❌", cmd, err.Error()))
		return "", err
	}

	output, ended := collectUntilIdle(lines, idle, axesByDevice)
	stopped := !ended || ctx.Err() != nil
	cancel()
	err = cmdExec.Wait()
	if stopped {
		c.logger.Info(fmt.Sprintf("(%s): stopped ✅", cmd))
		return output, nil
	}
	if err != nil {
		c.logger.Error(fmt.Sprintf("(%s): %s ❌", cmd, err.Error()))
	}
	return output, err
}

func startLineStream(ctx context.Context, cmd string) (*exec.Cmd, <-chan string, error) {
	cmdExec := exec.CommandContext(ctx, "bash", "-c", cmd)
	cmdExec.Stderr = os.Stderr
	cmdExec.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmdExec.Cancel = func() error {
		return syscall.Kill(-cmdExec.Process.Pid, syscall.SIGKILL)
	}
	stdout, err := cmdExec.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}
	err = cmdExec.Start()
	if err != nil {
		return nil, nil, err
	}

	lines := make(chan string)
	go readLines(ctx, stdout, lines)
	return cmdExec, lines, nil
}

func readLines(ctx context.Context, reader io.Reader, lines chan<- string) {
	defer close(lines)
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		select {
		case lines <- scanner.Text():
		case <-ctx.Done():
			return
		}
	}
}

func collectUntilIdle(
	lines <-chan string,
	idle time.Duration,
	axesByDevice map[string]touchAxes,
) (string, bool) {
	var output strings.Builder
	var timer = time.NewTimer(idle)
	defer timer.Stop()
	for {
		select {
		case line, ok := <-lines:
			if !ok {
				return output.String(), true
			}
			output.WriteString(line)
			output.WriteByte('\n')
			if isTouchLine(line, axesByDevice) {
				timer.Reset(idle)
			}
		case <-timer.C:
			return output.String(), false
		}
	}
}

func (c *cmdImpl) startInBackground(cmd string) (*backgroundProc, error) {
	if cmd == "" {
		return nil, fmt.Errorf("cmd is empty")
	}

	c.logger.Info("-------")
	cmdExec := exec.Command("bash", "-c", cmd)
	cmdExec.Stdin = os.Stdin
	cmdExec.Stderr = os.Stderr
	cmdExec.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	err := cmdExec.Start()
	if err != nil {
		c.logger.Error(fmt.Sprintf("(%s) error during starting in background: %s", cmd, err.Error()))
		return nil, err
	}
	c.logger.Info(fmt.Sprintf("(%s) Starting in background... ⏳", cmd))

	proc := &backgroundProc{cmd: cmdExec, done: make(chan struct{})}
	go proc.wait()
	return proc, nil
}

func (p *backgroundProc) wait() {
	_ = p.cmd.Wait()
	close(p.done)
}

func (c *cmdImpl) stopProc(proc *backgroundProc, grace time.Duration) {
	pid := proc.cmd.Process.Pid
	_ = syscall.Kill(-pid, syscall.SIGTERM)
	select {
	case <-proc.done:
		return
	case <-time.After(grace):
	}
	_ = syscall.Kill(-pid, syscall.SIGKILL)
	<-proc.done
}

func (c *cmdImpl) pidsMatching(pattern string) []int {
	output, err := exec.Command("pgrep", "-f", pattern).Output()
	if err != nil {
		return nil
	}
	var pids []int
	for _, field := range strings.Fields(string(output)) {
		pid, err := strconv.Atoi(field)
		if err == nil && pid > 0 {
			pids = append(pids, pid)
		}
	}
	return pids
}
