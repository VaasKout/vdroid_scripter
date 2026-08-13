package usecases

import (
	"android_vision_scripter/pkg/models"
	"fmt"
	"time"
)

// SessionUseCase ...
type SessionUseCase interface {
	StartSession(serial string, basePort int) bool
	CloseSession(serial string)
	GetPortsJSON(serial string) map[string]string
	GetSessionStatus(serial string) string
}

func (i *interactorImpl) StartSession(serial string, basePort int) bool {
	i.logger.Info(fmt.Sprintf("closing old connections for %s... ⏳", serial))
	i.CloseSession(serial)

	session := i.initPorts(basePort)
	if session == nil {
		return false
	}
	i.sessionsCache.Add(serial, *session)

	started := i.StartScrcpyServer(serial, session.ServerPort)
	if !started {
		return false
	}

	i.setScrcpyState(serial, true)
	go i.runSessionQueue(serial)
	return true
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

func (i *interactorImpl) GetPortsJSON(serial string) map[string]string {
	if result, ok := i.sessionsCache.Get(serial); ok {
		return map[string]string{
			"video_port":   fmt.Sprintf("%d", result.VideoPort),
			"cv_port":      fmt.Sprintf("%d", result.CVPort),
			"control_port": fmt.Sprintf("%d", result.ControlPort),
		}
	}
	return map[string]string{}
}

func (i *interactorImpl) initPorts(basePort int) *models.Session {
	var videoPort = basePort + 1
	var cvPort = videoPort + 1
	var controlPort = cvPort + 1

	var cacheMap = i.sessionsCache.GetMap()
	if len(cacheMap) == 0 {
		return &models.Session{
			ServerPort:  basePort,
			VideoPort:   videoPort,
			CVPort:      cvPort,
			ControlPort: controlPort,
			Status:      models.StatusIdle,
			DoneCh:      make(chan struct{}),
		}
	}

	var biggestPort = controlPort
	for _, conn := range cacheMap {
		if conn.ControlPort > biggestPort {
			biggestPort = conn.ControlPort
		}
	}

	serverPort := biggestPort + 1
	videoPort = serverPort + 1
	cvPort = videoPort + 1
	controlPort = cvPort + 1

	return &models.Session{
		ServerPort:  serverPort,
		VideoPort:   videoPort,
		CVPort:      cvPort,
		ControlPort: controlPort,
		Status:      models.StatusIdle,
		DoneCh:      make(chan struct{}),
	}
}

func (i *interactorImpl) GetSessionStatus(serial string) string {
	session, ok := i.sessionsCache.Get(serial)
	if !ok {
		return models.StatusClosed
	}
	return session.Status
}

func (i *interactorImpl) addStepsToQueue(serial string, steps []models.Step) {
	session, ok := i.sessionsCache.Get(serial)
	if !ok {
		return
	}
	session.Query = append(session.Query, steps...)
	i.sessionsCache.Add(serial, session)
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

		if len(session.Query) == 0 {
			time.Sleep(300 * time.Millisecond)
			continue
		}

		var step = session.Query[0]
		session.Query = session.Query[1:]
		session.Status = fmt.Sprintf(models.StatusRunningStep, step.ToString())
		i.sessionsCache.Add(serial, session)

		err := i.executeStep(serial, &step)
		if err != nil {
			i.logger.Error(err.Error())
			session.Status = err.Error()
			session.Query = []models.Step{}
			i.sessionsCache.Add(serial, session)
			return
		}

		if len(session.Query) == 0 {
			session.Status = models.StatusIdle
		}
		i.sessionsCache.Add(serial, session)
	}
}
