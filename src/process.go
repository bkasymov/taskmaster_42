package main

import (
	"strconv"
	"time"
)

// /***
// * Контроллер используется для остановки / запуска процессов
// * chans - каналы для передачи информации о том, какие каналы остановить или запустить
// */
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

type ProcessMap map[string][]*Process

type ProcExitCode int

// /***
// * Статусы процессов
// */
const (
	ProcRunning        = "running"
	ProcSetup          = "setup"
	ProcStopped        = "stopped"
	ProcCrashed        = "crashed"
	ProcDone           = "done"
	ProcUnableToStart  = "unable to start"
	ProcUnableToConfig = "unable to configure"
	ProcKilled         = "killed"
	ProcStopping       = "stopping"

	ProcExitOk ProcExitCode = iota
	ProcExitCrash
	ProcExitUnableToStart
	ProcExitKilled
	ProcExitConfErr
)

/*
Проходится по каждому конфигу. Для каждого конфига создает процесс и добавляет его в карту процессов в цикле
в зависимости от количества процессов, которые нужно запустить.
Сначала создаёт хранилище новых процессов в map, где ключ - имя процесса, а значение - характеристики процесса.
*/
func CreateProcesses(conf Config, num int) []*Process {
	processSlice := make([]*Process, num)
	for i := 0; i < num; i++ {
		process := &Process{
			Name:     conf.Name + " - " + strconv.Itoa(i),
			Conf:     conf,
			Status:   ProcStopped,
			Exit:     -1, // Exit -1 означает, что процесс еще не запускался
			Crashes:  0,
			Restarts: 0,
		}
		processSlice[i] = process
	}
	return processSlice
}

func SyncProcessesWithConfigs(configs map[string]Config) ProcessMap {
	pMap := ProcessMap{}
	for _, conf := range configs {
		processSlice := CreateProcesses(conf, conf.NumProcs)
		pMap[conf.Name] = processSlice
	}
	return pMap
}

//func (p Process) String() string {
//	return fmt.Sprintf("%s %d %s", p.Name, p.Pid, p.Status)
//}

///***
// * ProcessContainer - функция, которая запускает процесс и обрабатывает его результат (запускать в отдельной горутине)
// */
//
//func ProcessContainer(
//	ctx context.Context,
//	process *Process,
//	wg *sync.WaitGroup,
//	envlock *chan interface{},
//	doneChan chan *Process) {
//
//	// Когда процесс завершится, сообщаем о завершении в канал doneChan и уменьшаем счетчик wg
//	// defer - выполняется в конце функции (после return) и позволяет избежать дублирования кода в случае ошибки в функции (например, в случае panic)
//	defer wg.Done()
//	defer func() {
//		doneChan <- process
//	}()
//
//	restartsNum := process.Conf.StartRetries
//
//	for {
//		result := RunProcess(ctx, process, envlock)
//		handleProcessResult(process, result)
//
//		if restartsNum != 0 && (process.Conf.AutoRestart == "always" || (process.Conf.AutoRestart == "sometimes" && result == ProcExitUnableToStart)) {
//			logger.Println("Retrying process again:", process.Name)
//			if restartsNum > 0 {
//				restartsNum--
//			}
//			process.Restarts++
//		} else {
//			return
//		}
//	}
//}
//
//func RunProcess(ctx context.Context, process *Process, envlock *chan interface{}) interface{} {
//// TODO остановился тут. Разбор тут же.
//}
//
//func handleProcessResult(process *Process, result interface{}) {
//	switch result {
//	case ProcExitOk:
//		logger.Println(process.Name, "Ok")
//	case ProcExitCrash:
//		logger.Println(process.Name, "Crashed")
//		process.Crashes++
//	case ProcExitUnableToStart:
//		logger.Println(process.Name, "Unable to start")
//		process.Crashes++
//	case ProcExitKilled:
//		logger.Println(process.Name, "Killed by user")
//		return
//	case ProcExitConfErr:
//		logger.Println(process.Name, "Error configuring proc")
//	}
//}
