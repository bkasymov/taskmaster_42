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

	ctrl := controller{}
	fmt.Println(ctrl)
	ctrl.chans.init()
	waitchan := make(chan interface{})
	go ctrl.run(waitchan) // запускаем контроллер в отдельной горутине (потоке) и передаем канал для ожидания завершения работы контроллера (waitchan) в качестве аргумента функции run (см. controller.go)

	confs, err := updateConfig(ctrl.chans)
	fmt.Println(confs) // TODO убрать
	//if err != nil {
	//	logger.Println("Unable to load config:", err)
	//	return
	//}
	//
	//err = runUI(confs, ctrl.chans)
	//if err != nil {
	//	logger.Println("Unable to run visualizer. Exiting")
	//}
	//cleanupProcesses(ctrl.chans, waitchan)
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

/***
 * updateConfig is a function that updates the config file and returns the new config.
 */

func updateConfig(chans ProcChannels) (map[string][]*Process, error) {
	confs, err := UpdateConfig(configFile, map[string][]*Process{}, chans)
	return confs, err
}

// /***
// * cleanupProcesses is a function that cleans up processes and waits for them to finish before exiting the program.
// */
//func cleanupProcesses(chans *Channels, waitchan chan interface{}) {
//	logger.Println("Cleaning up processes")
//	close(chans.Killall)
//	<-waitchan
//}
