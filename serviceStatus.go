package main

import "sync"

const (
	// работает
	runServiceStatusInfo = "run"
	// перезапускается
	restartServiceStatusInfo = "restart"
	// останавливается
	stopServiceStatusInfo = "stop"
)

type serviceStatusInfo struct {
	sync.RWMutex
	state string
}

func newServiceStatusInfo() *serviceStatusInfo {
	return &serviceStatusInfo{}
}

func (s *serviceStatusInfo) run() {
	s.setState(runServiceStatusInfo)
}

func (s *serviceStatusInfo) isRun() bool {
	return s.getState() == runServiceStatusInfo
}

func (s *serviceStatusInfo) stop() {
	s.setState(stopServiceStatusInfo)
}

func (s *serviceStatusInfo) isStop() bool {
	return s.getState() == stopServiceStatusInfo
}

func (s *serviceStatusInfo) restart() {
	s.setState(restartServiceStatusInfo)
}

func (s *serviceStatusInfo) isRestart() bool {
	return s.getState() == restartServiceStatusInfo
}

func (s *serviceStatusInfo) setState(state string) {
	s.Lock()
	s.state = state
	s.Unlock()
}

func (s *serviceStatusInfo) getState() string {
	s.RLock()
	defer s.RUnlock()
	return s.state
}
