package bashcmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Path properties
const (
	ScrCpyDefaultPath = "/data/local/tmp/scrcpy-server.jar"
	ScrcpyProcessName = "scrcpy"
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

	return c.executeInBackground(command)
}

func (c *cmdImpl) scrcpyPids(serial string) []string {
	var filteredPids []string
	pids := c.pidsOfProcess(ScrcpyProcessName)
	if len(pids) > 0 {
		psList := c.psAuxList(ScrcpyProcessName)
		for _, ps := range psList {
			for _, pid := range pids {
				var hasSerial = strings.Contains(ps, serial)
				var hasPid = strings.Contains(ps, pid)

				if hasSerial && hasPid {
					filteredPids = append(filteredPids, pid)
				}
			}
		}
	}
	return filteredPids
}

// KillScrcpy ...
func (c *cmdImpl) KillScrcpy(serial string) {
	pids := c.scrcpyPids(serial)
	for _, pid := range pids {
		pidInt, err := strconv.Atoi(pid)
		if err != nil {
			fmt.Printf("Failed to convert pid to int: %s\n", pid)
			continue
		}
		proc, _ := os.FindProcess(pidInt)
		proc.Kill()
		c.logger.Info("-------")
		c.logger.Info(fmt.Sprintf("Killing process %d for %s... ⏳", pidInt, serial))
		proc.Wait()
		c.logger.Info(fmt.Sprintf("Killed process %d for %s ✅", pidInt, serial))
		c.logger.Info("-------")
	}
}
