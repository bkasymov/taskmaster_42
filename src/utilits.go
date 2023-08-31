package main

import (
	"bytes"
	"fmt"
	"sort"
)

func getSortedKeys(p ProcessMap) []string {
	keys := make([]string, 0, len(p))
	for key := range p {
		keys = append(keys, key)
	}
	sort.Strings(keys) // Вместо использования sort.Slice можно использовать sort.Strings
	return keys
}

func formatProcesses(procs []*Process) string {
	var b bytes.Buffer
	for _, proc := range procs {
		b.WriteString(proc.String())
		b.WriteString("\n")
	}
	return b.String()
}

func (p ProcessMap) String() string {
	var b bytes.Buffer
	keys := getSortedKeys(p)
	for _, key := range keys {
		b.WriteString(key)
		b.WriteString(":\n")
		b.WriteString(formatProcesses(p[key]))
	}
	return b.String()
}
func (p Process) String() string {
	return fmt.Sprintf("%s %d %s", p.Name, p.Pid, p.Status)
}
