package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	serverBinaryName    = "vdroid-scripter"
	serverBinaryEnv     = "VDROID_BIN"
	serverStartTimeout  = 15 * time.Second
	serverStopTimeout   = 5 * time.Second
	deviceSettleTimeout = 10 * time.Second
	serverPollInterval  = time.Second
	pingTimeout         = time.Second
)

var serverBinaryFallbacks = []string{
	"/usr/local/bin/vdroid-scripter",
	"/opt/homebrew/bin/vdroid-scripter",
}

var errNotVdroid = errors.New("not a vdroid server")

func (c *apiClient) recoverServer() (bool, error) {
	c.startMu.Lock()
	defer c.startMu.Unlock()

	err := c.ping()
	if err == nil {
		return false, nil
	}
	if errors.Is(err, errNotVdroid) {
		return false, err
	}
	if err := c.launchServer(); err != nil {
		return false, err
	}
	return true, nil
}

func (c *apiClient) ping() error {
	resp, err := c.pingClient.Get(c.baseURL + "/ping")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK || !strings.Contains(string(body), `"ok"`) {
		return fmt.Errorf(
			"%s answers but %w (GET /ping returned %d): another program holds the port",
			c.baseURL, errNotVdroid, resp.StatusCode,
		)
	}
	return nil
}

func (c *apiClient) launchServer() error {
	if !c.isLoopback() {
		return fmt.Errorf(
			"vdroid server at %s is not reachable and is not local, so it cannot be started from here",
			c.baseURL,
		)
	}

	binary, err := serverBinary()
	if err != nil {
		return err
	}
	logPath, logFile, err := serverLogFile()
	if err != nil {
		return err
	}

	cmd := exec.Command(binary)
	if home, err := os.UserHomeDir(); err == nil {
		cmd.Dir = home
	}
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	err = cmd.Start()
	logFile.Close()
	if err != nil {
		return fmt.Errorf("could not start %s: %w", binary, err)
	}
	go cmd.Wait()

	deadline := time.Now().Add(serverStartTimeout)
	for time.Now().Before(deadline) {
		if c.ping() == nil {
			c.waitForDevices()
			return nil
		}
		time.Sleep(serverPollInterval)
	}
	return fmt.Errorf(
		"started %s but nothing answered at %s within %s, see %s",
		binary, c.baseURL, serverStartTimeout, logPath,
	)
}

func (c *apiClient) waitForDevices() {
	deadline := time.Now().Add(deviceSettleTimeout)
	for time.Now().Before(deadline) {
		if c.hasDevices() {
			return
		}
		time.Sleep(serverPollInterval)
	}
}

func (c *apiClient) hasDevices() bool {
	body, err := c.send(http.MethodGet, "/devices", nil)
	if err != nil {
		return false
	}
	var response struct {
		Devices []json.RawMessage `json:"devices"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return false
	}
	return len(response.Devices) > 0
}

func (c *apiClient) stopServer() (bool, error) {
	if !c.isLoopback() {
		return false, fmt.Errorf("vdroid server at %s is not local, stop it on its own machine", c.baseURL)
	}

	pids, err := serverPids()
	if err != nil {
		return false, err
	}
	if len(pids) == 0 {
		if c.ping() == nil {
			return false, fmt.Errorf(
				"%s answers but no %s process is running (started with go run?), stop it by hand",
				c.baseURL, serverBinaryName,
			)
		}
		return false, nil
	}

	signalAll(pids, syscall.SIGTERM)
	if waitUntilGone(serverStopTimeout) {
		return true, nil
	}
	signalAll(pids, syscall.SIGKILL)
	if waitUntilGone(serverPollInterval) {
		return true, nil
	}
	return false, fmt.Errorf("%s did not exit after SIGTERM and SIGKILL", serverBinaryName)
}

func serverPids() ([]int, error) {
	output, err := exec.Command("pgrep", "-x", serverBinaryName).Output()
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("pgrep failed: %w", err)
	}

	pids := []int{}
	for _, line := range strings.Fields(string(output)) {
		pid, err := strconv.Atoi(line)
		if err != nil {
			continue
		}
		pids = append(pids, pid)
	}
	return pids, nil
}

func signalAll(pids []int, sig syscall.Signal) {
	for _, pid := range pids {
		process, err := os.FindProcess(pid)
		if err != nil {
			continue
		}
		process.Signal(sig)
	}
}

func waitUntilGone(timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		pids, err := serverPids()
		if err == nil && len(pids) == 0 {
			return true
		}
		time.Sleep(serverPollInterval)
	}
	pids, err := serverPids()
	return err == nil && len(pids) == 0
}

func (c *apiClient) isLoopback() bool {
	parsed, err := url.Parse(c.baseURL)
	if err != nil {
		return false
	}
	host := parsed.Hostname()
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func serverBinary() (string, error) {
	if custom := os.Getenv(serverBinaryEnv); custom != "" {
		return custom, nil
	}
	if found, err := exec.LookPath(serverBinaryName); err == nil {
		return found, nil
	}
	for _, candidate := range serverBinaryFallbacks {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	return "", fmt.Errorf(
		"%s is not installed (run install.sh) and %s is not set",
		serverBinaryName, serverBinaryEnv,
	)
}

func serverLogFile() (string, *os.File, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", nil, err
	}
	logDir := filepath.Join(cacheDir, "vdroid_scripter", "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return "", nil, err
	}
	logPath := filepath.Join(logDir, "server.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return "", nil, err
	}
	return logPath, logFile, nil
}
