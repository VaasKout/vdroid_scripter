package bashcmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

const (
	AdbStateDevice      = "device"
	trackDevicesCommand = "adb track-devices -l"
	adbBlockHeaderSize  = 4
	modelAttribute      = "model:"
	deviceAttribute     = "device:"
)

var adbEntryAttributes = []string{"usb:", "product:", modelAttribute, deviceAttribute, "transport_id:"}

type AdbEntry struct {
	Serial string
	State  string
	Model  string
	Device string
}

func (c *cmdImpl) TrackDevices() (<-chan []AdbEntry, error) {
	_ = exec.Command("adb", "start-server").Run()
	cmdExec := exec.Command("adb", "track-devices", "-l")
	cmdExec.Stderr = os.Stderr
	cmdExec.SysProcAttr = &syscall.SysProcAttr{Pdeathsig: syscall.SIGTERM}
	stdout, err := cmdExec.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmdExec.Start(); err != nil {
		return nil, err
	}
	c.logger.Info(fmt.Sprintf("(%s): tracking devices... ⏳", trackDevicesCommand))

	entries := make(chan []AdbEntry)
	go readAdbEntries(cmdExec, stdout, entries)
	return entries, nil
}

func readAdbEntries(cmdExec *exec.Cmd, stdout io.Reader, entries chan<- []AdbEntry) {
	defer close(entries)
	reader := bufio.NewReader(stdout)
	for {
		block, err := readAdbBlock(reader)
		if err != nil {
			break
		}
		entries <- parseAdbEntries(block)
	}
	_ = cmdExec.Wait()
}

func readAdbBlock(reader *bufio.Reader) (string, error) {
	header := make([]byte, adbBlockHeaderSize)
	if _, err := io.ReadFull(reader, header); err != nil {
		return "", err
	}
	length, err := strconv.ParseUint(string(header), 16, 32)
	if err != nil {
		return "", err
	}
	block := make([]byte, length)
	if _, err := io.ReadFull(reader, block); err != nil {
		return "", err
	}
	return string(block), nil
}

func parseAdbEntries(block string) []AdbEntry {
	entries := []AdbEntry{}
	for _, line := range strings.Split(block, "\n") {
		entry, ok := parseAdbEntry(line)
		if ok {
			entries = append(entries, entry)
		}
	}
	return entries
}

func parseAdbEntry(line string) (AdbEntry, bool) {
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return AdbEntry{}, false
	}
	entry := AdbEntry{Serial: fields[0]}
	state := []string{}
	for _, field := range fields[1:] {
		key, value, attribute := adbAttribute(field)
		if !attribute {
			state = append(state, field)
			continue
		}
		switch key {
		case modelAttribute:
			entry.Model = value
		case deviceAttribute:
			entry.Device = value
		}
	}
	entry.State = strings.Join(state, " ")
	return entry, true
}

func adbAttribute(field string) (string, string, bool) {
	for _, key := range adbEntryAttributes {
		if strings.HasPrefix(field, key) {
			return key, strings.TrimPrefix(field, key), true
		}
	}
	return "", "", false
}
