package bashcmd

import (
	"fmt"
	"regexp"
	"syscall"
	"time"
)

// Path properties
const (
	ScrCpyDefaultPath = "/data/local/tmp/scrcpy-server.jar"
	scrcpyStopGrace   = 3 * time.Second
)

// ScrcpyCmdConfig ...
type ScrcpyCmdConfig struct {
	Version       string
	TunnelForward bool
	LogLevel      string
	Audio         bool
	Cleanup       bool
	RawStream     bool
	MaxSize       string
	Codec         string
	MaxFPS        string
	VideoBitrate  string
	StayAwake     bool
}

// SetDefault ...
func (s *ScrcpyCmdConfig) SetDefault(version string) {
	if s == nil {
		s = &ScrcpyCmdConfig{}
	}
	s.Version = version
	s.TunnelForward = true
	s.LogLevel = "verbose"
	s.Audio = false
	s.Cleanup = true
	s.RawStream = false
	// s.MaxSize = "1920"
	// s.Codec = "h264"
	s.MaxFPS = "120"
	// s.VideoBitrate = "40000"
	s.StayAwake = true
}

// ScrCpy ...
type ScrCpy interface {
	ConnectToScrcpy(serial string, config *ScrcpyCmdConfig) error
	KillScrcpy(serial string)
}

// ConnectToScrCpy ...
func (c *cmdImpl) ConnectToScrcpy(serial string, config *ScrcpyCmdConfig) error {
	if config == nil {
		return fmt.Errorf("failed to launch scrcpy for %s: config is nil", serial)
	}
	var commandFormat = "adb -s %s shell CLASSPATH=%s app_process / com.genymobile.scrcpy.Server %s"
	var command = fmt.Sprintf(commandFormat, serial, ScrCpyDefaultPath, config.Version)

	if config.TunnelForward {
		command += " tunnel_forward=true"
	}

	if config.Codec != "" {
		command += fmt.Sprintf(" video_codec=%s", config.Codec)
	}

	if config.LogLevel != "" {
		command += fmt.Sprintf(" log_level=%s", config.LogLevel)
	}

	if !config.Audio {
		command += " audio=false"
	}

	if !config.Cleanup {
		command += " cleanup=false"
	}

	if config.RawStream {
		command += " raw_stream=true"
	}

	if config.MaxSize != "" {
		command += fmt.Sprintf(" max_size=%s", config.MaxSize)
	}

	if config.MaxFPS != "" {
		command += fmt.Sprintf(" max_fps=%s", config.MaxFPS)
	}

	if config.VideoBitrate != "" {
		command += fmt.Sprintf(" video_bit_rate=%s", config.VideoBitrate)
	}

	if config.StayAwake {
		command += " stay_awake=true"
	}

	proc, err := c.startInBackground(command)
	if err != nil {
		return err
	}
	c.scrcpyMu.Lock()
	previous := c.scrcpyProcs[serial]
	c.scrcpyProcs[serial] = proc
	c.scrcpyMu.Unlock()
	if previous != nil {
		c.stopProc(previous, scrcpyStopGrace)
	}
	return nil
}

func scrcpyLaunchPattern(serial string) string {
	return "^adb -s " + regexp.QuoteMeta(serial) + " shell CLASSPATH=" + regexp.QuoteMeta(ScrCpyDefaultPath)
}

func (c *cmdImpl) KillScrcpy(serial string) {
	c.scrcpyMu.Lock()
	proc := c.scrcpyProcs[serial]
	delete(c.scrcpyProcs, serial)
	c.scrcpyMu.Unlock()
	if proc != nil {
		c.logger.Info(fmt.Sprintf("stopping scrcpy shell %d for %s... ⏳", proc.cmd.Process.Pid, serial))
		c.stopProc(proc, scrcpyStopGrace)
		c.logger.Info(fmt.Sprintf("scrcpy shell for %s stopped ✅", serial))
	}

	for _, pid := range c.pidsMatching(scrcpyLaunchPattern(serial)) {
		c.logger.Info(fmt.Sprintf("killing stray scrcpy shell %d for %s... ⏳", pid, serial))
		_ = syscall.Kill(pid, syscall.SIGKILL)
	}
}
