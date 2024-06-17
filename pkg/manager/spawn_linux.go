package manager

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/njayp/jobber/pkg/manager/cgroups"
	"github.com/njayp/jobber/pkg/pb"
)

// Spawn creates a job-level cgroup,
// moves to the cgroup, and spawns the process.
// It then moves back to the "jobber" cgroup
// and waits for the process to exit, and
// logs its exit code.
func Spawn(id, exe string, args ...string) error {
	// load user cgroup
	ucg, err := cgroups.LoadCGroup(userCGPath)
	if err != nil {
		return err
	}

	// create target cgroup
	cg, err := ucg.NewChildCGroup(id)
	if err != nil {
		return err
	}
	err = cg.SetLimits()
	if err != nil {
		return err
	}

	// add self to cgroup
	err = cg.AddProc(fmt.Sprint(os.Getpid()))
	if err != nil {
		return err
	}

	// create cmd
	cmd := exec.Command(exe, args...)
	// direct output streams to files
	err = mutateCmdOut(cmd, id)
	if err != nil {
		return err
	}
	// run cmd as linux user
	err = mutateCmdUser(cmd, userName)
	if err != nil {
		return err
	}

	// spawn process in cgroup
	err = cmd.Start()
	if err != nil {
		return err
	}

	// move self back to safe cgroup
	jcg, err := cgroups.LoadCGroup(jobberCGPath)
	if err != nil {
		return err
	}
	err = jcg.AddProc(fmt.Sprint(os.Getpid()))
	if err != nil {
		return err
	}

	// wait for cmd to exit so that we can log exit code
	err = cmd.Wait()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode := exitError.ExitCode()
			return logExitCode(id, exitCode)
		}
		return err
	}
	return logExitCode(id, 0)
}

// mutateCmdOut sets stdout and stderr of a cmd to files in the tmp dir
func mutateCmdOut(cmd *exec.Cmd, id string) error {
	stdoutPath := outFilePath(id, pb.StreamSelect_Stdout)

	// Create the directories if they do not exist
	err := os.MkdirAll(filepath.Dir(stdoutPath), os.ModePerm)
	if err != nil {
		return err
	}

	// direct stdout to file
	sof, err := os.Create(stdoutPath)
	if err != nil {
		return err
	}
	cmd.Stdout = sof

	// direct stderr to file
	stderrPath := outFilePath(id, pb.StreamSelect_Stderr)
	sef, err := os.Create(stderrPath)
	if err != nil {
		return err
	}
	cmd.Stderr = sef

	return nil
}

// mutateCmdUser makes the cmd run as user
func mutateCmdUser(cmd *exec.Cmd, username string) error {
	// Look up the user
	usr, err := user.Lookup(username)
	if err != nil {
		return err
	}

	// Convert the user's UID and GID to integers
	uid, err := strconv.Atoi(usr.Uid)
	if err != nil {
		return err
	}
	gid, err := strconv.Atoi(usr.Gid)
	if err != nil {
		return err
	}

	// Set the UID and GID
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Credential: &syscall.Credential{
			Uid: uint32(uid),
			Gid: uint32(gid),
		},
	}

	return nil
}
