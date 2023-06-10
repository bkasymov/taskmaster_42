package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
	"sort"
	"syscall"
)

type Config struct {
	Name         string // name of program
	Signal       os.Signal
	Cmd          string            `yaml:"cmd"`      // binary to run
	Args         []string          `yaml:"args"`     // list of args
	NumProcs     int               `yaml:"numprocs"` // number of processes
	Umask        int               `yaml:"umask"`    // int representing permissions
	WorkingDir   string            `yaml:"workingdir"`
	AutoStart    bool              `yaml:"autostart"`    // true/false (default: true)
	AutoRestart  string            `yaml:"autorestart"`  // always/never/sometimes (defult: never)
	ExitCodes    []int             `yaml:"exitcodes"`    // expected exit codes (default: 0)
	StartRetries int               `yaml:"startretries"` // times to retry if sometimes exit (default: 0) (-1 for infinite)
	StartTime    int               `yaml:"starttime"`    // time to start app
	StopSignal   string            `yaml:"stopsignal"`   // signal to kill
	StopTime     int               `yaml:"stoptime"`     // time until mean kill
	Stdin        string            `yaml:"stdin"`        // file read as stdin
	Stdout       string            `yaml:"stdout"`       // stdout redirect file
	Stderr       string            `yaml:"stderr"`       // stderr redirect file
	Env          map[string]string `yaml:"env"`          // map of env vars
}

// Взяты наиболее часто используемые сигналы из https://golang.org/pkg/syscall/#pkg-constants
var signals = map[string]os.Signal{
	"ABRT":   syscall.SIGABRT,
	"ALRM":   syscall.SIGALRM,
	"BUS":    syscall.SIGBUS,
	"CHLD":   syscall.SIGCHLD,
	"CONT":   syscall.SIGCONT,
	"FPE":    syscall.SIGFPE,
	"HUP":    syscall.SIGHUP,
	"ILL":    syscall.SIGILL,
	"INT":    syscall.SIGINT,
	"IO":     syscall.SIGIO,
	"IOT":    syscall.SIGIOT,
	"KILL":   syscall.SIGKILL,
	"PIPE":   syscall.SIGPIPE,
	"PROF":   syscall.SIGPROF,
	"QUIT":   syscall.SIGQUIT,
	"SEGV":   syscall.SIGSEGV,
	"STOP":   syscall.SIGSTOP,
	"SYS":    syscall.SIGSYS,
	"TERM":   syscall.SIGTERM,
	"TRAP":   syscall.SIGTRAP,
	"TSTP":   syscall.SIGTSTP,
	"TTIN":   syscall.SIGTTIN,
	"TTOU":   syscall.SIGTTOU,
	"URG":    syscall.SIGURG,
	"USR1":   syscall.SIGUSR1,
	"USR2":   syscall.SIGUSR2,
	"VTALRM": syscall.SIGVTALRM,
	"WINCH":  syscall.SIGWINCH,
	"XCPU":   syscall.SIGXCPU,
	"XFSZ":   syscall.SIGXFSZ,
}

//ParsePrograms()
//|
//|-- readFile()
//|
//|-- unmarshalYaml()
//|
//|-- parsePrograms()
//|   |
//|   |-- parseConfig()
//|   |   |
//|   |   |-- setDefaultValues()
//|   |   |   |
//|   |   |   |-- setDefaultValuesIfNotPresent()
//|   |

func GetSyscallSignal(sig string) (os.Signal, error) {
	signal := signals[sig]
	if signal == nil {
		return nil, fmt.Errorf("invalid process signal: %s", sig)

	}
	return signal, nil
}

func (c *Config) String() string {
	format := `
		Name: %s
		Cmd: %s
		Args: %s
		NumProcs: %d
		Umask: %d
		StartRetries: %d
		----------------
		WorkingDir: %s
		AutoStart: %t
		AutoRestart: %s
		ExitCodes: %d
		StartTime: %d
		StopSignal: %s
		StopTime: %d
		----------------
		Stdin: %s
		Stdout: %s
		Stderr: %s
		----------------
		Env: %s
		`
	return fmt.Sprintf(format, c.Name, c.Cmd, c.Args, c.NumProcs, c.Umask, c.StartRetries, c.WorkingDir, c.AutoStart, c.AutoRestart, c.ExitCodes, c.StartTime, c.StopSignal, c.StopTime, c.Stdin, c.Stdout, c.Stderr, c.Env)
}

// readFile(file string) ([]byte, error): Эта функция читает содержимое файла, переданного в аргументе file, и возвращает его как байтовый массив. В случае ошибки при чтении файла она вернет эту ошибку.
func readFile(file string) ([]byte, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// unmarshalYaml(data []byte) (map[interface{}]interface{}, error): Эта функция преобразует байтовый массив data, представляющий данные в формате YAML, в карту Go, где ключ и значение являются обобщенными интерфейсами. В случае ошибки при преобразовании данных она вернет эту ошибку.
func unmarshalYaml(data []byte) (map[interface{}]interface{}, error) {
	ymap := make(map[interface{}]interface{})
	err := yaml.Unmarshal(data, &ymap)
	if err != nil {
		return nil, err
	}
	return ymap, nil
}

// Эта функция принимает карту, возвращаемую unmarshalYaml, и парсит каждую пару ключ-значение, представляющую конфигурацию программы. Она возвращает карту, где ключ - это имя программы, а значение - структура конфигурации.
func parsePrograms(ymap map[interface{}]interface{}) (map[string]Config, error) {
	configs := make(map[string]Config)
	configMap, ok := ymap["programs"].(map[interface{}]interface{})
	if !ok {
		return nil, fmt.Errorf("programs not found")
	}

	for key, value := range configMap {
		config, err := parseConfig(value)
		if err != nil {
			return configs, err
		}
		name, ok := key.(string)
		if !ok {
			return configs, fmt.Errorf("invalid name: %s", name)
		}
		config.Name = name
		configs[config.Name] = config
	}
	return configs, nil
}

// parseConfig parses individual configuration
func parseConfig(value interface{}) (Config, error) {
	config := Config{}
	data, err := yaml.Marshal(value)
	if err != nil {
		return config, err
	}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return config, err
	}
	err = setDefaultValues(data, &config)
	if err != nil {
		return config, err
	}
	return config, nil
}

// Эта функция принимает байтовый массив data, представляющий конфигурацию программы в формате YAML, и ссылку на структуру конфигурации config. Она устанавливает значения по умолчанию для полей конфигурации, которые не указаны в данных YAML.
func setDefaultValues(data []byte, config *Config) error {
	configMap2 := make(map[interface{}]interface{})
	err := yaml.Unmarshal(data, &configMap2)
	if err != nil {
		return err
	}

	setDefaultValuesIfNotPresent(configMap2, config)

	return nil
}

// Эта функция принимает карту, представляющую конфигурацию программы, и ссылку на структуру конфигурации. Она устанавливает значения по умолчанию для полей конфигурации, которые не указаны в карте.
func setDefaultValuesIfNotPresent(configMap2 map[interface{}]interface{}, config *Config) {
	if ok := configMap2["autostart"]; ok == nil {
		config.AutoStart = true
	}
	if ok := configMap2["autorestart"]; ok == nil {
		config.AutoRestart = "never"
	}
	if ok := configMap2["umask"]; ok == nil {
		config.Umask = 022
	}
	if ok := configMap2["stoptime"]; ok == nil {
		config.StopTime = 1
	}
	if ok := configMap2["workingdir"]; ok == nil {
		config.WorkingDir = "./"
	}
	if ok := configMap2["stopsignal"]; ok == nil {
		config.StopSignal = "INT"
	}
	config.Signal, _ = GetSyscallSignal(config.StopSignal)
	if ok := configMap2["exitcodes"]; ok == nil {
		config.ExitCodes = []int{0}
	}
	sort.Ints(config.ExitCodes)
	if config.AutoRestart == "" {
		config.AutoRestart = "sometimes"
	}
	if config.NumProcs <= 0 {
		config.NumProcs = 1
	}
}

// Это основная функция, которая объединяет все предыдущие. Она читает файл конфигурации программ, преобразует его в карту, парсит каждую конфигурацию программы и возвращает карту конфигураций. Если в любой точке происходит ошибка, она вернет эту ошибку.
func ParsePrograms(file string) (map[string]Config, error) {
	data, err := readFile(file)
	if err != nil {
		return nil, err
	}

	ymap, err := unmarshalYaml(data)
	if err != nil {
		return nil, err
	}

	configs, err := parsePrograms(ymap)
	if err != nil {
		return nil, err
	}

	return configs, nil
}

func UpdateConfig(
	filename string,
	last ProcessMap,
	pc ProcChannels) (ProcessMap, error) {

	new, err := ParsePrograms(filename)
	if err != nil {
		logger.Println("Error parsing config file:", err)
		return last, err
	}
	newProcesses := ConfigToProcess(new)
	logger.Println("Apply new configs")
	return nil, nil

}
