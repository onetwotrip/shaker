package main

import (
	"github.com/foxdalas/shaker/pkg/shaker"

	"github.com/bamzi/jobrunner"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

var AppVersion = "unknown"
var AppGitCommit = ""
var AppGitState = ""
var stopCh chan struct{}

func Version() string {
	version := AppVersion
	if len(AppGitCommit) > 0 {
		version += "-"
		version += AppGitCommit[0:8]
	}
	if len(AppGitState) > 0 && AppGitState != "clean" {
		version += "-"
		version += AppGitState
	}
	return version
}

func main() {
	jobrunner.Start()

	s := shaker.New(Version())
	go s.Init()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	signal := <-c
	logger := logrus.WithField("signal", signal.String())
	logger.Debug("received signal")
	Stop()

}

func Stop() {
	logrus.Info("shutting things down")
	stopCh := make(chan struct{})
	close(stopCh)
}
