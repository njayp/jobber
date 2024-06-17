package manager

import (
	"bytes"
	"context"
	"fmt"
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
		file.WriteString(fmt.Sprint(i))
		i += 1
	}

	t.Run("inotify", func(t *testing.T) {
		// set a watch
		ch := make(chan struct{})
		go inotify(context.Background(), filename, ch)

		// wait for watch to setup before writing to file
		time.Sleep(time.Second)

		// check that we are notified when writing to file x5
		for range [5]struct{}{} {
			// write to file
			write()

			select {
			case <-time.After(time.Second):
				t.Error("took too long to receive notify")
			case <-ch:
			}
		}

		// check that we are not notified when not writing
		select {
		case <-time.After(time.Second):
		case <-ch:
			t.Error("received bad notify")
		}
	})

	t.Run("copyfollow", func(t *testing.T) {
		var buf bytes.Buffer
		ctx, cancel := context.WithCancel(context.Background())

		// copyfollow should read the current file into buffer immediately
		go copyFollow(ctx, filename, &buf)
		// wait for copy
		time.Sleep(time.Second)
		if buf.String() != "12345" {
			t.Errorf("buf value was wrong: %s", buf.String())
		}

		// write to file. copyfollow should write it to buffer
		write()
		time.Sleep(time.Second)
		if buf.String() != "123456" {
			t.Errorf("buf value was wrong: %s", buf.String())
		}

		// copyfollow should not write after cancel
		cancel()
		write()
		time.Sleep(time.Second)
		if buf.String() != "123456" {
			t.Errorf("buf value was wrong: %s", buf.String())
		}
	})
}
