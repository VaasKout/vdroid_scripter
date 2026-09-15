package usecases

import (
	"android_vision_scripter/pkg/core/file"
	"android_vision_scripter/pkg/models"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Recording errors
var (
	ErrDeviceBusy          = errors.New("device is busy, wait for the session to become idle")
	ErrRecordingInProgress = errors.New("device is recording a gesture, wait for it to finish")
)

// SessionUseCase ...
type SessionUseCase interface {
	StartSession(serial string, basePort int) bool
	CloseSession(serial string)
	CloseAllSessions()
	GetPortsJSON(serial string) map[string]string
	GetSessionStatus(serial string) string
	RecordAction(serial string, name string, basePort int) (bool, error)
}

func (i *interactorImpl) StartSession(serial string, basePort int) bool {
	i.logger.Info(fmt.Sprintf("closing old connections for %s... ⏳", serial))
	i.CloseSession(serial)

	session := i.initPorts(basePort)
	if session == nil {
		return false
	}

	started := i.StartScrcpyServer(serial, session.ServerPort)
	if !started {
		return false
	}

	i.sessionsCache.Add(serial, *session)
	i.setScrcpyState(serial, true)
	go i.runSessionQueue(serial)
	return true
}

func (i *interactorImpl) ensureSessionIsRunning(serial string, basePort int) error {
	if _, ok := i.sessionsCache.Get(serial); ok {
		return nil
	}

	if !i.StartSession(serial, basePort) {
		return fmt.Errorf("couldn't start scrcpy server for %s", serial)
	}
	go i.scrcpy.ReadVideoStream(serial, nil)

	deadline := time.Now().Add(time.Duration(models.FirstFrameTimeout) * time.Second)
	for time.Now().Before(deadline) {
		mat, err := i.scrcpy.GetMatFromLastFrame(serial, true)
		if err != nil || mat == nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		mat.Close()
		return nil
	}
	return fmt.Errorf("no video frame received from %s", serial)
}

func (i *interactorImpl) CloseSession(serial string) {
	i.logger.Info(
		fmt.Sprintf(
			"closing scrcpy connection for %s... 🛑",
			serial,
		),
	)
	i.scrcpy.CloseScrcpyServer(serial)
	if connection, ok := i.sessionsCache.Get(serial); ok {
		close(connection.DoneCh)
		i.sessionsCache.Delete(serial)
	}
	i.setScrcpyState(serial, false)
}

func (i *interactorImpl) CloseAllSessions() {
	for serial := range i.sessionsCache.GetMap() {
		i.CloseSession(serial)
	}
}

func (i *interactorImpl) GetPortsJSON(serial string) map[string]string {
	if result, ok := i.sessionsCache.Get(serial); ok {
		return map[string]string{
			"video_port":   fmt.Sprintf("%d", result.VideoPort),
			"control_port": fmt.Sprintf("%d", result.ControlPort),
		}
	}
	return map[string]string{}
}

func (i *interactorImpl) initPorts(basePort int) *models.Session {
	var videoPort = basePort + 1
	var controlPort = videoPort + 1

	var cacheMap = i.sessionsCache.GetMap()
	if len(cacheMap) == 0 {
		return &models.Session{
			ServerPort:  basePort,
			VideoPort:   videoPort,
			ControlPort: controlPort,
			Status:      models.StatusIdle,
			DoneCh:      make(chan struct{}),
		}
	}

	var biggestPort = controlPort
	for _, conn := range cacheMap {
		biggestPort = max(conn.ControlPort, biggestPort)
	}

	serverPort := biggestPort + 1
	videoPort = serverPort + 1
	controlPort = videoPort + 1

	return &models.Session{
		ServerPort:  serverPort,
		VideoPort:   videoPort,
		ControlPort: controlPort,
		Status:      models.StatusIdle,
		DoneCh:      make(chan struct{}),
	}
}

func (i *interactorImpl) runSessionQueue(serial string) {
	defer i.logger.Info(fmt.Sprintf("queue loop closed for %s... 🛑", serial))
	for {
		session, ok := i.sessionsCache.Get(serial)
		if !ok {
			return
		}

		select {
		case <-session.DoneCh:
			return
		default:
		}

		step, found := i.popNextStep(serial)
		if !found {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		err := i.executeStep(serial, &step)
		if err != nil {
			i.logger.Error(err.Error())
			i.failSessionQueue(serial, err)
			continue
		}

		i.finishSessionStep(serial)
	}
}

func (i *interactorImpl) GetSessionStatus(serial string) string {
	session, ok := i.sessionsCache.Get(serial)
	if !ok {
		return models.StatusClosed
	}
	return session.Status
}

func (i *interactorImpl) addStepsToQueue(serial string, steps []models.Step) bool {
	session, ok := i.sessionsCache.Get(serial)
	if !ok {
		return false
	}
	session.Query = append(session.Query, steps...)
	i.sessionsCache.Add(serial, session)
	return true
}

func (i *interactorImpl) popNextStep(serial string) (models.Step, bool) {
	session, ok := i.sessionsCache.Get(serial)
	if !ok || len(session.Query) == 0 || session.Status == models.StatusRecording {
		return models.Step{}, false
	}

	step := session.Query[0]
	session.Query = session.Query[1:]
	session.Status = fmt.Sprintf(models.StatusRunningStep, step.ToString())
	i.sessionsCache.Add(serial, session)
	return step, true
}

func (i *interactorImpl) failSessionQueue(serial string, err error) {
	session, ok := i.sessionsCache.Get(serial)
	if !ok {
		return
	}
	session.Status = err.Error()
	session.Query = nil
	i.sessionsCache.Add(serial, session)
}

func (i *interactorImpl) finishSessionStep(serial string) {
	session, ok := i.sessionsCache.Get(serial)
	if !ok || len(session.Query) != 0 {
		return
	}
	session.Status = models.StatusIdle
	i.sessionsCache.Add(serial, session)
}

func (i *interactorImpl) RecordAction(serial string, name string, basePort int) (bool, error) {
	serial = strings.TrimSpace(serial)
	name = strings.TrimSpace(name)
	if serial == "" {
		return false, errors.New(SerialIsEmptyError)
	}
	if !file.ValidName(name) {
		return false, errors.New("invalid action name")
	}
	if err := i.ensureSessionIsRunning(serial, basePort); err != nil {
		return false, err
	}
	if !i.startRecording(serial) {
		return false, ErrDeviceBusy
	}
	defer i.finishRecording(serial)

	i.logger.Info(fmt.Sprintf(
		"recording %s on %s for %ds... ⏳", name, serial, models.RecordDurationSeconds,
	))
	width, height, err := i.scrcpy.GetScreenSize(serial)
	if err != nil {
		return false, err
	}
	action, err := i.cmd.RecordTouches(
		serial,
		time.Duration(models.RecordDurationSeconds)*time.Second,
		width,
		height,
	)
	if err != nil {
		return false, err
	}

	action.Name = name
	if action.IsEmpty() {
		i.logger.Info(fmt.Sprintf("nothing recorded for %s on %s", name, serial))
		return false, nil
	}
	if !i.SaveAction(action) {
		return false, fmt.Errorf("couldn't save action %s", name)
	}
	i.logger.Info(fmt.Sprintf("recorded %s with %d events ✅", name, len(action.Events)))
	return true, nil
}

func (i *interactorImpl) startRecording(serial string) bool {
	session, ok := i.sessionsCache.Get(serial)
	if !ok || len(session.Query) != 0 {
		return false
	}
	if session.Status == models.StatusRecording || models.IsRunningStatus(session.Status) {
		return false
	}
	session.Status = models.StatusRecording
	i.sessionsCache.Add(serial, session)
	return true
}

func (i *interactorImpl) finishRecording(serial string) {
	session, ok := i.sessionsCache.Get(serial)
	if !ok || session.Status != models.StatusRecording {
		return
	}
	session.Status = models.StatusIdle
	i.sessionsCache.Add(serial, session)
}
