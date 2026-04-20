package main

import (
	"log/slog"
	"os"
	"os/exec"
	"time"

	"golang.org/x/sys/windows/svc"
)

type service struct{ config Config }

func (m *service) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	file, errr := os.OpenFile(m.config.Errorlogs, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if errr != nil {
		errno = 1
		return

	}
	defer file.Close()

	logger := slog.New(slog.NewJSONHandler(file, nil))
	changes <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}
	cmd := exec.Command(m.config.Execpath, m.config.Args...)

	if m.config.Path != "" {
		cmd.Dir = m.config.Path
	}

	cmd.Stderr = file
	logger.Info("Starting child process")
	err := cmd.Start()

	if err != nil {
		logger.Error("Failed to start child process", "error", err)
		return
	}
	done := make(chan error)
	go func() {
		done <- cmd.Wait()
	}()

loop:
	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				cmd.Process.Kill()
				break loop

			}
		case err := <-done:
			if err != nil {
				logger.Error("child process exited with error",
					"error", err.Error(),
					"exit_code", cmd.ProcessState.ExitCode(),
				)
				if m.config.Restart {
					logger.Info("Restarting in", "delay seconds", m.config.Restartdelay)
					delay := m.config.Restartdelay
					<-time.After(time.Duration(delay) * time.Second)
					cmd = exec.Command(m.config.Execpath, m.config.Args...)
					if err := cmd.Start(); err != nil {
						logger.Error("Failed to restart child process", "error", err)
					}
					done = make(chan error)
					go func() {
						done <- cmd.Wait()
					}()
				}
			} else {
				logger.Info("Child process exited successfully")
			}
		}
	}
	return
}
