package cgroups

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type CGroup struct {
	path string
}

func LoadCGroup(path string) (*CGroup, error) {
	// check if cgroup folder exists
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("cgroup does not exist: %w", err)
		}
		return nil, err
	}

	return &CGroup{path: path}, nil
}

func (c *CGroup) NewChildCGroup(name string) (*CGroup, error) {
	path := filepath.Join(c.path, name)
	err := os.MkdirAll(path, 0755)
	if err != nil {
		return nil, err
	}

	return &CGroup{path: path}, nil
}

func (c *CGroup) SetLimits() error {
	// 10 %
	err := c.setCPUMax(10)
	if err != nil {
		return err
	}

	// 10MB
	err = c.setMemMax(10)
	if err != nil {
		return err
	}

	// 1MB/s
	err = c.setIOMax(1, 1)
	if err != nil {
		return err
	}

	// 100
	return c.setPidsMax(100)
}

func (c *CGroup) Path() string {
	return c.path
}

func (c *CGroup) AddProc(pid string) error {
	return c.setCGroupValue("cgroup.procs", pid)
}

func (c *CGroup) Kill() error {
	return c.setCGroupValue("cgroup.kill", "1")
}

func (c *CGroup) SetSubtreeControl() error {
	return c.setCGroupValue("cgroup.subtree_control", "+cpu +memory +io +pids")
}

func (c *CGroup) ReadCGroupValue(file string) ([]byte, error) {
	fullPath := filepath.Join(c.path, file)
	return os.ReadFile(fullPath)
}

func (c *CGroup) Pids() ([]string, error) {
	data, err := c.ReadCGroupValue("cgroup.procs")
	if err != nil {
		return nil, err
	}
	return strings.Fields(string(data)), nil
}
