package main

import (
	"context"
	"sync"
	"time"
)

type Process struct {
	Name         string
	Conf         Config
	Pid          int
	Status       string
	Crashes      int
	Restarts     int
	Exit         int
	StartTime    time.Time
	StopTime     time.Time
	StopDuration time.Duration
}

type ProcExitCode int

const (
	ProcRunning = "running"
	ProcSetup = "setup"
	ProcStopped = "stopped"
	ProcCrashed = "crashed"
	ProcDone = "done"
	ProcUnableToStart = "unable to start"
	ProcUnableToConfig = "unable to configure"
	ProcKilled = "killed"
	ProcStopping = "stopping"

	ProcExitOk  ProcExitCode = iota
	ProcExitCrash
	ProcExitUnableToStart
	ProcExitKilled
	ProcExitConfErr
}

func ProcessContainer(ctx context.Context, process *Process, wg *sync.WaitGroup,
	envlock *chan interface{}, doneChan chan *Process) {
	// Defer wg.Done mean that the function will be executed at the end of the function
	defer wg.Done()

	// When process has done its job, it will send itself to doneChan
	defer func() {
		doneChan <- process
	}()

	restartsNum := process.Conf.StartRetries

	for {
		result := RunProcess(ctx, process, envlock)
		handleProcessResult(process, result)

		if restartsNum != 0 && (process.Conf.AutoRestart == "always" || (process.Conf.AutoRestart == "sometimes" && result == ProcExitUnableToStart)) {
			logger.Println("Retrying process again:", process.Name)
			if restartsNum > 0 {
				restartsNum--
			}
			process.Restarts++
		} else {
			return
		}
	}
}

func RunProcess(ctx context.Context, process *Process, envlock *chan interface{}) interface{} {
// TODO остановился тут. Разбор тут же.
}

func handleProcessResult(process *Process, result interface{}) {
	switch result {
	case ProcExitOk:
		logger.Println(process.Name, "Ok")
	case ProcExitCrash:
		logger.Println(process.Name, "Crashed")
		process.Crashes++
	case ProcExitUnableToStart:
		logger.Println(process.Name, "Unable to start")
		process.Crashes++
	case ProcExitKilled:
		logger.Println(process.Name, "Killed by user")
		return
	case ProcExitConfErr:
		logger.Println(process.Name, "Error configuring proc")
	}
}
