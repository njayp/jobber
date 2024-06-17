package manager

import (
	"context"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/njayp/jobber/pkg/manager/cgroups"
	"github.com/njayp/jobber/pkg/pb"
)

// Start calls Spawn in a new process
func Start(req *pb.StartRequest) (*pb.StartResponse, error) {
	exe, err := os.Executable()
	if err != nil {
		return nil, err
	}

	id := uuid.NewString()
	args := append([]string{"spawn", id}, req.CmdString...)

	go func() {
		// wait for Spawn to complete, and log err
		err = exec.Command(exe, args...).Run()
		if err != nil {
			log.Println("spawning process encountered an error: ", err)
		}
	}()

	return &pb.StartResponse{Id: id}, nil
}

// Stop kills job with id
func Stop(req *pb.StopRequest) (*pb.StopResponse, error) {
	path := filepath.Join(userCGPath, req.Id)
	cg, err := cgroups.LoadCGroup(path)
	if err != nil {
		return nil, err
	}

	err = cg.Kill()
	if err != nil {
		return nil, err
	}

	return &pb.StopResponse{}, nil
}

// Status of job with id
func Status(req *pb.StatusRequest) (*pb.StatusResponse, error) {
	path := filepath.Join(userCGPath, req.Id)
	cg, err := cgroups.LoadCGroup(path)
	if err != nil {
		return nil, err
	}

	state := pb.State_Running

	// check if job is not running
	pids, err := cg.Pids()
	if len(pids) == 0 {
		state = pb.State_Exited

		// check if job was killed
		k, err := isKilled(req.Id)
		if err != nil {
			return nil, err
		}
		if k {
			state = pb.State_Killed
		}
	}

	return &pb.StatusResponse{State: state}, nil
}

// Stream streams "stdout.txt" or "stderr.txt"
func Stream(ctx context.Context, req *pb.StreamRequest, w io.Writer) error {
	path := outFilePath(req.Id, req.StreamSelect)
	return copyFollow(ctx, path, w)
}
