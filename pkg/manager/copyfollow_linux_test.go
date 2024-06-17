package manager

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"testing"
	"time"
)

func TestCopyFollow(t *testing.T) {
	filename := "inotify-test.txt"

	// Create or open the file
	file, err := os.Create(filename)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	i := 1
	write := func() {
		_, err = file.WriteString(fmt.Sprint(i))
		if err != nil {
			t.Fatal(err)
		}
		i += 1
	}

	t.Run("copyfollow", func(t *testing.T) {
		// buffer to store data
		var buf bytes.Buffer
		// pipe blocks until data
		pr, pw := io.Pipe()
		// read pipe into buffer
		tr := io.TeeReader(pr, &buf)
		ctx, cancel := context.WithCancel(context.Background())
		write()

		// copyfollow should read the current file into buffer immediately
		chErr := make(chan error)
		go func() {
			err := copyFollow(ctx, filename, pw)
			chErr <- err
		}()

		// pipe into buf
		p := make([]byte, 1024)
		tr.Read(p)
		if buf.String() != "1" {
			t.Errorf("buf value was wrong: %s", buf.String())
		}

		// write to file. copyfollow should write it to pipe
		write()
		// pipe into buf
		tr.Read(p)
		if buf.String() != "12" {
			t.Errorf("buf value was wrong: %s", buf.String())
		}

		// copyfollow should return after cancel
		cancel()
		select {
		case <-time.After(time.Second):
			t.Error("copyfollow did not return")
		case err := <-chErr:
			if err != nil {
				t.Errorf("copyfollow returned with err: %s", err)
			}
		}
	})
}
