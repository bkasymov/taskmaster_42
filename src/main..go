package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

var logger *log.Logger
var configFile string

func main() {
	flag.Usage = func() {
		fmt.Println("usage:", os.Args[0], "[options] config.yaml")
		flag.PrintDefaults()
	}
	logfile := setupLogging()

	defer logfile.Close()

	err := parseFlags()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// init controller and channels for communication between controller and processes (see controller.go)
	ctrl := controller{}
	ctrl.chans.init()

	waitchan := make(chan interface{}) // Создаем канал для ожидания завершения работы контроллера (waitchan)

	//TODO конфиги обновляются. Теперь надо запустить процессы
	go ctrl.run(waitchan) // Запускаем контроллер в отдельной горутине (потоке) и передаем канал для ожидания завершения работы контроллера (waitchan) в качестве аргумента функции run (см. controller.go)
	// Обновляем конфигурацию либо читаем в первый раз (см. update.go)
	processes, err := ReloadConfig(configFile, map[string][]*Process{}, ctrl.chans)
	if err != nil {
		logger.Println("Unable to load config:", err)
		return
	}
	err = runGUI(processes, ctrl.chans)

	fmt.Println(processes)
}

/***
 * setupLogging is a function that sets up the logging for the program.
 * It returns the log file.
 * The default log file is /tmp/taskmaster.log
 * The log file is opened in append mode.
 * The log file is created if it does not exist.
 * The log file has permissions 0644.
 * The log file is closed in main.
 */

func setupLogging() *os.File {
	var logname string
	flag.StringVar(&logname, "logfile", "/tmp/taskmaster.log", "log file")
	flag.Parse()
	logfile, err := os.OpenFile(logname, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	logger = log.New(logfile, "taskmaster: ", log.Lshortfile|log.Ltime)
	return logfile
}

/***
 * parseFlags is a function that parses the flags provided to the program.
 * It returns an error if the number of arguments is not 1.
 * The argument is the config.yaml file.
 */
func parseFlags() error {
	if flag.NArg() != 1 {
		return fmt.Errorf("config.yaml file not provided")
	}
	configFile = flag.Arg(0)
	return nil
}

// /***
// * cleanupProcesses is a function that cleans up processes and waits for them to finish before exiting the program.
// */
//func cleanupProcesses(chans *Channels, waitchan chan interface{}) {
//	logger.Println("Cleaning up processes")
//	close(chans.Killall)
//	<-waitchan
//}

//
//err = runUI(confs, ctrl.chans)
//if err != nil {
//	logger.Println("Unable to run visualizer. Exiting")
//}
//cleanupProcesses(ctrl.chans, waitchan)
