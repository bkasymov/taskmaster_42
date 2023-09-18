package main

import "reflect"

func ReloadConfig(
	filename string,
	old ProcessMap,
	pc ProcChannels) (ProcessMap, error) {

	new, err := ParsePrograms(filename)
	if err != nil {
		logger.Println("Error parsing config file:", err)
		return old, err
	}
	newProcesses := SyncProcessesWithConfigs(new)
	logger.Println("Applying new configs...")
	processNewConfigs(newProcesses, old, pc)
	removeOldProcesses(old, pc)
	return newProcesses, nil
}

func removeOldProcesses(old ProcessMap, pc ProcChannels) {
	for _, slices := range old {
		for _, v := range slices {
			pc.oldPros <- v
		}
	}
}

func processNewConfigs(newProcesses ProcessMap, old ProcessMap, pc ProcChannels) {
	for key, newSlices := range newProcesses {
		lastSlices, ok := old[key]
		if !ok {
			handleNewProcesses(key, newSlices, pc)
		} else {
			handleExistingProcesses(key, newSlices, lastSlices, pc)
			delete(old, key)
		}
	}
}

// handleExistingProcesses handles existing processes in config file and restarts them if they are set to auto start in config.
func handleExistingProcesses(key string, newSlices []*Process, lastSlices []*Process, p ProcChannels) {
	if reflect.DeepEqual(newSlices[0].Conf, lastSlices[0].Conf) {
		return
	}
	for _, v := range newSlices {
		p.oldPros <- v
		if v.Conf.AutoStart {
			logger.Println("Relaunching process:", v.Name)
			p.newPros <- v
		}
	}
}

// handleNewProcesses handles new processes in config file and starts them if they are set to auto start in config.
func handleNewProcesses(key string, slices []*Process, p ProcChannels) {
	for _, v := range slices {
		if v.Conf.AutoStart {
			p.newPros <- v
		}
	}
}
