// Package bashcmd ...
package bashcmd

import (
	"android_vision_scripter/internal/filesdb"
	"android_vision_scripter/pkg/logger"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

// CmdAPI ...
type CmdAPI interface {
	AdbAPI
	ScrCpy

	ExecuteCommand(serial string) (string, error)
}
type cmdImpl struct {
	logger  *logger.Logger
	filesDB filesdb.FilesDB
}

// New instance of CmdAPI
func New(
	filesDB filesdb.FilesDB,
	logger *logger.Logger,
) CmdAPI {
	return &cmdImpl{
		logger:  logger,
		filesDB: filesDB,
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

func (c *cmdImpl) executeWithTimeout(cmd string, timeout time.Duration) (string, error) {
	if cmd == "" {
		return "", fmt.Errorf("cmd is empty")
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmdExec := exec.CommandContext(ctx, "bash", "-c", cmd)
	cmdExec.Stderr = os.Stderr
	cmdExec.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmdExec.Cancel = func() error {
		return syscall.Kill(-cmdExec.Process.Pid, syscall.SIGKILL)
	}

	c.logger.Info("-------")
	c.logger.Info(fmt.Sprintf("(%s): Start for %s... ⏳", cmd, timeout))
	result, err := cmdExec.Output()
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		c.logger.Info(fmt.Sprintf("(%s): stopped after %s ✅", cmd, timeout))
		return string(result), nil
	}
	if err != nil {
		c.logger.Error(fmt.Sprintf("(%s): %s ❌", cmd, err.Error()))
	}
	return string(result), err
}

func (c *cmdImpl) executeInBackground(cmd string) error {
	if cmd == "" {
		return fmt.Errorf("cmd is empty")
	}

	c.logger.Info("-------")
	cmdExec := exec.Command("bash", "-c", cmd)
	cmdExec.Stdin = os.Stdin
	cmdExec.Stderr = os.Stderr
	err := cmdExec.Start()
	if err != nil {
		c.logger.Error(fmt.Sprintf("(%s) error during starting in background: %s", cmd, err.Error()))
		return err
	}
	c.logger.Info(fmt.Sprintf("(%s) Starting in background... ⏳", cmd))
	return nil
}

func (c *cmdImpl) pidsOfProcess(name string) []string {
	result, err := c.ExecuteCommand(fmt.Sprintf("pgrep -f %s", name))
	if err != nil {
		return []string{}
	}
	var formattedResult = strings.TrimSpace(result)
	return strings.Split(formattedResult, "\n")
}

func (c *cmdImpl) psAuxList(filter string) []string {
	var cmd = "ps aux"
	if filter != "" {
		cmd = fmt.Sprintf("%s | grep %s", cmd, filter)
	}
	result, err := c.ExecuteCommand(cmd)
	if err != nil {
		return []string{}
	}
	var formattedResult = strings.TrimSpace(result)
	return strings.Split(formattedResult, "\n")
}
