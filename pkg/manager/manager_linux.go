package manager

import "github.com/njayp/jobber/pkg/manager/cgroups"

// use NewManager to initialize
type Manager struct{}

// ensures that initCGroups is run
func NewManager() (*Manager, error) {
	err := initCGroups()
	if err != nil {
		return nil, err
	}

	return &Manager{}, nil
}

// initCGroups must be called before any other
// actions are called. It can be called more than once.
// It sets up the cgroup tree so that jobs can be added
// as children of the user node.
func initCGroups() error {
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
