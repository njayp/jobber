package manager

import (
	"context"
	"fmt"
	"io"
	"os"
	"syscall"
	"unsafe"
)

func copyFollow(ctx context.Context, filepath string, w io.Writer) error {
	// Open the file
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	// make watcher that will signal when file is modified
	// Create an inotify instance
	inotifyFd, err := syscall.InotifyInit()
	if err != nil {
		return fmt.Errorf("initializing inotify: %w\n", err)
	}
	defer syscall.Close(inotifyFd)

	// Add a watch for the file
	watchFd, err := syscall.InotifyAddWatch(inotifyFd, filepath, syscall.IN_MODIFY)
	if err != nil {
		return fmt.Errorf("adding watch: %w\n", err)
	}

	// prevents syscall.Read from blocking after ctx is cancelled
	go func() {
		defer syscall.InotifyRmWatch(inotifyFd, uint32(watchFd))
		<-ctx.Done()
	}()

	// read file into writer
	_, err = io.Copy(w, file)
	if err != nil {
		return err
	}

	// Buffer to receive events
	var buf [syscall.SizeofInotifyEvent * 10]byte

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			// Read events
			n, err := syscall.Read(inotifyFd, buf[:])
			if err != nil {
				return err
			}

			// Process each event
			var offset uint32
			for offset < uint32(n) {
				event := (*syscall.InotifyEvent)(unsafe.Pointer(&buf[offset]))
				// event can also be syscall.IN_DELETE or syscall.IN_ATTRIB
				if event.Mask&syscall.IN_MODIFY == syscall.IN_MODIFY {
					// read file into writer, then wait for next read
					_, err = io.Copy(w, file)
					if err != nil {
						return err
					}
					break
				}
				offset += syscall.SizeofInotifyEvent + event.Len
			}
		}
	}
}
