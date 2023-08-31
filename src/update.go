package main

func UpdateConfig(
	filename string,
	last ProcessMap,
	pc ProcChannels) (ProcessMap, error) {

	new, err := ParsePrograms(filename)
	if err != nil {
		logger.Println("Error parsing config file:", err)
		return last, err
	}
	newProcesses := SyncProcessesWithConfigs(new)
	logger.Println("Applying new configs...")
	// TODO остановился тут

	return newProcesses, nil
}
