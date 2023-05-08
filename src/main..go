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
	ctrl.chans.init()
	waitchan := make(chan interface{})
	go ctrl.run(waitchan)

	confs, err := updateConfig(ctrl.chans)
	if err != nil {
		logger.Println("Unable to load config:", err)
		return
	}

	err = runUI(confs, ctrl.chans)
	if err != nil {
		logger.Println("Unable to run visualizer. Exiting")
	}
	cleanupProcesses(ctrl.chans, waitchan)
}

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

func parseFlags() error {
	if flag.NArg() != 1 {
		return fmt.Errorf("config.yaml file not provided")
	}
	configFile = flag.Arg(0)
	return nil
}

func updateConfig(chans *Channels) (map[string][]*Process, error) {
	confs, err := UpdateConfig(configFile, map[string][]*Process{}, chans)
	return confs, err
}

func cleanupProcesses(chans *Channels, waitchan chan interface{}) {
	logger.Println("Cleaning up processes")
	close(chans.Killall)
	<-waitchan
}
