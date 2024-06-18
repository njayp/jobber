package cgroups

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

func (c *CGroup) setCGroupValue(file, value string) error {
	fullPath := filepath.Join(c.path, file)
	// TODO maybe append to file
	return os.WriteFile(fullPath, []byte(value), 0755)
}

func (c *CGroup) setCPUMax(max int) error {
	return c.setCGroupValue("cpu.max", fmt.Sprintf("%v 100000", max*1000))
}

func (c *CGroup) setMemMax(max int) error {
	return c.setCGroupValue("memory.max", fmt.Sprintf("%vM", max))
}

func (c *CGroup) setIOMax(rMax, wMax int) error {
	return c.setCGroupValue("io.max", fmt.Sprintf("1:0 rbps=%v wbps=%v", rMax*1024*1024, wMax*1024*1024))
}

func (c *CGroup) setPidsMax(max int) error {
	return c.setCGroupValue("pids.max", strconv.Itoa(max))
}
