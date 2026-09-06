package main

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

const (
	serverBinaryName   = "vdroid-scripter"
	serverBinaryEnv    = "VDROID_BIN"
	serverStartTimeout = 15 * time.Second
	serverPollInterval = 200 * time.Millisecond
)

var serverBinaryFallbacks = []string{
	"/usr/local/bin/vdroid-scripter",
	"/opt/homebrew/bin/vdroid-scripter",
}

func isConnectionRefused(err error) bool {
	return errors.Is(err, syscall.ECONNREFUSED)
}

func (c *apiClient) startServer() error {
	c.startMu.Lock()
	defer c.startMu.Unlock()

	if c.ping() == nil {
		return nil
	}
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
			return nil
		}
		time.Sleep(serverPollInterval)
	}
	return fmt.Errorf(
		"started %s but nothing answered at %s within %s, see %s",
		binary, c.baseURL, serverStartTimeout, logPath,
	)
}

func (c *apiClient) ping() error {
	resp, err := c.client.Get(c.baseURL + "/ping")
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ping returned %d", resp.StatusCode)
	}
	return nil
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
