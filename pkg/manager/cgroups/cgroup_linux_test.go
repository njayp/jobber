package cgroups

import (
	"errors"
	"io/fs"
	"path/filepath"
	"testing"
)

func TestCGroups(t *testing.T) {
	rootPath := "/sys/fs/cgroup"

	t.Run("load does not exist", func(t *testing.T) {
		path := filepath.Join(rootPath, "does-not-exist")
		_, err := LoadCGroup(path)
		if !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("wrong err received: %s", err)
		}
	})
}
