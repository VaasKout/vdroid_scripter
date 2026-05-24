// Package scrcpy ...
package scrcpy

import (
	"android_vision_scripter/config"
	"android_vision_scripter/internal/bashcmd"
	"android_vision_scripter/internal/cv"
	"android_vision_scripter/internal/filesdb"
	"android_vision_scripter/pkg/core/cache"
	"android_vision_scripter/pkg/core/file"
	"android_vision_scripter/pkg/core/network"
	"android_vision_scripter/pkg/logger"
	"fmt"
	"net"
	"net/http"
	"path/filepath"
	"time"
)

// Scrcpy contst
const (
	ScrcpyTag         = "scrcpy"
	ScrcpyLinkFormat  = "https://github.com/Genymobile/scrcpy/releases/download/v%s/scrcpy-server-v%s"
	ScrcpyFileFormat  = "scrcpy-v%s"
	LocalTCPUrlFormat = "tcp://127.0.0.1:%d"
)

// CV type consts
const (
	NoCV = iota
	CVRects
)

// Attempts to find rectangle and establish scrcpy connection on start
const (
	Attempts     = 3
	ConnAttempts = 4
)

// Data ...
type Data struct {
	Port        int
	VideoConn   net.Conn
	ControlConn net.Conn
	Data        *DecoderData
}

// Scrcpy ...
type Scrcpy interface {
	StartScrcpyServer(
		serial string,
		port int,
	) string

	CloseScrcpyServer(serial string)
	Connection
}

type scrcpyImpl struct {
	cmd         bashcmd.CmdAPI
	cv          cv.API
	filesDB     filesdb.FilesDB
	props       *config.ScrcpyProps
	logAPI      *logger.Logger
	scrcpyCache cache.Cache[*Data]
}

// New scrcpy instance
func New(
	cmd bashcmd.CmdAPI,
	cv cv.API,
	fileDB filesdb.FilesDB,
	props *config.ScrcpyProps,
	logAPI *logger.Logger,
) Scrcpy {
	var scrcpyCache = cache.NewSafeCache[*Data]()

	var s = &scrcpyImpl{
		cmd:         cmd,
		cv:          cv,
		filesDB:     fileDB,
		props:       props,
		logAPI:      logAPI,
		scrcpyCache: scrcpyCache,
	}

	err := s.downloadScrcpyServer()
	if err != nil {
		panic(err)
	}
	return s
}

func (s *scrcpyImpl) downloadScrcpyServer() error {
	dir := s.filesDB.CreateScrcpyDir()
	scrcpyFileName := fmt.Sprintf(ScrcpyFileFormat, s.props.ScrcpyVersion)
	scrcpyFilePath := filepath.Join(dir, scrcpyFileName)

	if file.Exists(scrcpyFilePath) {
		s.logAPI.Info(fmt.Sprintf("found scrcpy file: %s", scrcpyFilePath))
		return nil
	}
	client := network.New(s.logAPI)
	var downloadLink = fmt.Sprintf(ScrcpyLinkFormat, s.props.ScrcpyVersion, s.props.ScrcpyVersion)
	request := &network.HTTPRequest{
		URL:     downloadLink,
		Method:  http.MethodGet,
		LogBody: true,
		LogReq:  true,
	}
	return client.DownloadFile(request, scrcpyFilePath)
}

func (s *scrcpyImpl) StartScrcpyServer(
	serial string,
	port int,
) string {
	dir := s.filesDB.CreateScrcpyDir()
	scrcpyFileName := fmt.Sprintf(ScrcpyFileFormat, s.props.ScrcpyVersion)
	scrcpyFilePath := filepath.Join(dir, scrcpyFileName)

	err := s.cmd.PushFile(serial, scrcpyFilePath, bashcmd.ScrCpyDefaultPath)
	if err != nil {
		s.logAPI.Error(err.Error())
		return ""
	}
	err = s.cmd.ForwardTCPPort(serial, port, ScrcpyTag)
	if err != nil {
		s.logAPI.Error(err.Error())
		return ""
	}

	var scrcpyCmdConfig = &bashcmd.ScrcpyCmdConfig{}
	scrcpyCmdConfig.SetDefault(s.props.ScrcpyVersion)
	err = s.cmd.ConnectToScrcpy(serial, scrcpyCmdConfig)
	if err != nil {
		s.logAPI.Error(err.Error())
		return ""
	}

	time.Sleep(1 * time.Second)
	for range ConnAttempts { // it takes too long to start scrcpy on some cheap devices
		err := s.initConnections(serial, port)
		if err != nil {
			s.logAPI.Error(err.Error())
			time.Sleep(1 * time.Second)
			continue
		}
		break
	}
	return fmt.Sprintf(LocalTCPUrlFormat, port)
}

func (s *scrcpyImpl) CloseScrcpyServer(serial string) {
	if result, ok := s.scrcpyCache.Get(serial); ok {
		result.VideoConn.Close()
		result.ControlConn.Close()
		result.Data.Free()
		s.scrcpyCache.Delete(serial)
	}
	s.cmd.KillScrcpy(serial)
}
