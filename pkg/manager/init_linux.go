package manager

import (
	"github.com/njayp/jobber/pkg/manager/cgroups"
)

// InitCGroups must be called only once before any other
// actions are called
func InitCGroups() error {
	root, err := cgroups.LoadCGroup(rootPath)
	if err != nil {
		return err
	}

	// move init process to jobber cgroup
	jobber, err := root.NewChildCGroup(jobberName)
	if err != nil {
		return err
	}
	err = jobber.AddProc("1")
	if err != nil {
		return err
	}

	// now we can set controls
	err = root.SetSubtreeControl()
	if err != nil {
		return err
	}

	// make `jobs` node and set controls
	jobs, err := root.NewChildCGroup(parentName)
	if err != nil {
		return err
	}
	err = jobs.SetSubtreeControl()
	if err != nil {
		return err
	}

	// make user nodes and set controls
	user, err := jobs.NewChildCGroup(userName)
	if err != nil {
		return err
	}
	err = user.SetSubtreeControl()
	if err != nil {
		return err
	}

	return nil
}
