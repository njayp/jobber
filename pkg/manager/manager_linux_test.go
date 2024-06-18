package manager

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/njayp/jobber/pkg/manager/cgroups"
	"github.com/njayp/jobber/pkg/pb"
)

func TestManager(t *testing.T) {
	// run init
	m, err := NewManager()

	// setup cgroup and check that init process has been moved
	t.Run("init cgroups", func(t *testing.T) {
		if err != nil {
			// cancel other tests if init fails
			t.Fatalf("failed to init gcroups: %s", err)
		}

		// check that cgroup exists
		path := filepath.Join(rootPath, "jobber")
		c, err := cgroups.LoadCGroup(path)
		if err != nil {
			t.Fatalf("jobber cgroup not found: %s", err)
		}
		// check that pid 1 has been added
		pids, err := c.Pids()
		if err != nil {
			t.Fatalf("failed to get pids: %s", err)
		}
		for _, pid := range pids {
			if pid == "1" {
				return
			}
		}
		t.Fatalf("Failed to find pid 1: %+v", pids)
	})

	// now that cgroup controllers are enabled, spawn a process
	// NOTE spawn moves testing process,
	// tests with Spawn cannot be run in parallel
	t.Run("spawn process", func(t *testing.T) {
		id := "cgroup-id"
		go Spawn(id, "watch", "date")

		// give cmd time to start
		time.Sleep(time.Second)

		// load cgroup created by spawn
		cg, err := cgroups.LoadCGroup(filepath.Join(userCGPath, id))
		if err != nil {
			t.Fatalf("failed to load cgroup: %s", err)
		}

		// check that processes are in cgroup
		pids, err := cg.Pids()
		if err != nil {
			t.Errorf("failed to get pids: %s", err)
		}
		if len(pids) == 0 {
			t.Error("no pids in target cgroup")
		}

		t.Run("test cgroups tree and limits", func(t *testing.T) {
			// assume cgroups works, so test whether limits were set
			// check that pids.max has been set to 100
			data, err := cg.ReadCGroupValue("pids.max")
			if err != nil {
				t.Error(err)
			}
			maxPids := strings.TrimSpace(string(data))
			if maxPids != "100" {
				t.Errorf("did not set correct pids max: %s", maxPids)
			}
			// TODO test other limits
		})

		t.Run("test status and kill", func(t *testing.T) {
			// move testing process before we call Status/Kill
			jcg, err := cgroups.LoadCGroup(jobberCGPath)
			if err != nil {
				t.Error(err)
			}
			err = jcg.AddProc(fmt.Sprint(os.Getpid()))
			if err != nil {
				t.Error(err)
			}

			// test Status on a running process
			resp, err := m.Status(&pb.StatusRequest{Id: id})
			if err != nil {
				t.Errorf("failed to get Status: %s", err)
			}
			if resp.State != pb.State_Running {
				t.Errorf("got wrong state: %s", resp.State)
			}

			// test Stop cgroup
			_, err = m.Stop(&pb.StopRequest{Id: id})
			if err != nil {
				t.Errorf("failed to kill cgroup: %s", err)
			}
			// wait a second for processes get killed
			time.Sleep(time.Second)

			// test Status on a killed process
			resp, err = m.Status(&pb.StatusRequest{Id: id})
			if err != nil {
				t.Errorf("failed to get Status: %s", err)
			}
			if resp.State != pb.State_Killed {
				t.Errorf("got wrong state: %s", resp.State)
			}
		})
	})

	t.Run("test process permissions", func(t *testing.T) {
		// spawn a cmd that needs user permissions
		id := "permissions-cgroup-id"
		go Spawn(id, "mkdir", filepath.Join(userCGPath, "should-not-exist"))

		// move testing process before we check status
		jcg, err := cgroups.LoadCGroup(jobberCGPath)
		if err != nil {
			t.Fatal(err)
		}
		err = jcg.AddProc(fmt.Sprint(os.Getpid()))
		if err != nil {
			t.Fatal(err)
		}

		// give process time to run
		time.Sleep(time.Second)

		// test Status on a exited process
		resp, err := m.Status(&pb.StatusRequest{Id: id})
		if err != nil {
			t.Errorf("failed to get Status: %s", err)
		}
		if resp.State != pb.State_Exited {
			t.Errorf("got wrong state: %s", resp.State)
		}

		// test that process ran as user and failed to execute
		var buf bytes.Buffer
		// test stream stderr
		go m.Stream(context.Background(), &pb.StreamRequest{Id: id, StreamSelect: pb.StreamSelect_Stderr}, &buf)
		// give time to io.Copy
		time.Sleep(time.Second)
		text := strings.TrimSpace(buf.String())
		if text != "mkdir: can't create directory '/sys/fs/cgroup/jobs/nobody/should-not-exist': Permission denied" {
			t.Errorf("stderr text not expected: %s", text)
		}
	})
}
