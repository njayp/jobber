package manager

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/njayp/jobber/pkg/pb"
)

const (
	rootPath         = "/sys/fs/cgroup"
	jobberName       = "jobber"
	parentName       = "jobs"
	userName         = "nobody"
	exitCodeFileName = "exitcode.txt"
)

var jobberCGPath = filepath.Join(rootPath, jobberName)
var userCGPath = filepath.Join(rootPath, parentName, userName)

// outFilePath returns the filepath to the selected file within
// the tmp dir. Since these files are streamed, pb.StreamSelect
// is used as the file selector
func outFilePath(id string, si pb.StreamSelect) string {
	var filename string
	switch si {
	case pb.StreamSelect_Stdout:
		filename = "stdout.txt"
	case pb.StreamSelect_Stderr:
		filename = "stderr.txt"
	// negative numbers for files not exposed in proto
	case -1:
		filename = exitCodeFileName
	default:
		filename = "unspecified"
	}
	return filepath.Join(os.TempDir(), "jobber", userName, id, filename)
}

func logExitCode(id string, code int) error {
	path := outFilePath(id, -1)
	// dir already exists
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(fmt.Sprint(code))
	if err != nil {
		return err
	}

	return nil
}

func isKilled(id string) (bool, error) {
	path := outFilePath(id, -1)
	data, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	text := string(data)
	// exec reports exit code -1 if process is killed via signal
	if text == "-1" {
		return true, nil
	}
	return false, nil
}
