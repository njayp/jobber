package manager

import (
	"context"
	"fmt"
	"io"
	"log"
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
	ch := make(chan struct{})
	go func() {
		err := inotify(ctx, filepath, ch)
		if err != nil {
			log.Println("inotify error: ", err)
		}
	}()

	// read file into writer
	_, err = io.Copy(w, file)
	if err != nil {
		return err
	}

	// wait for inotify or context cancelled
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ch:
			// read file into writer
			_, err := io.Copy(w, file)
			if err != nil {
				return err
			}
		}
	}
}

func inotify(ctx context.Context, filepath string, ch chan<- struct{}) error {
	// Create an inotify instance
	inotifyFd, err := syscall.InotifyInit()
	if err != nil {
		return fmt.Errorf("Error initializing inotify: %v\n", err)
	}
	defer syscall.Close(inotifyFd)

	// Add a watch for the file
	watchFd, err := syscall.InotifyAddWatch(inotifyFd, filepath, syscall.IN_MODIFY)
	if err != nil {
		return fmt.Errorf("Error adding watch: %v\n", err)
	}

	// prevents syscall.Read from blocking forever
	go func() {
		defer syscall.InotifyRmWatch(inotifyFd, uint32(watchFd))
		<-ctx.Done()
	}()

	// Buffer to receive events
	var buf [syscall.SizeofInotifyEvent * 10]byte

	for {
		// Read events
		n, err := syscall.Read(inotifyFd, buf[:])
		if err != nil {
			return fmt.Errorf("Error reading inotify events: %v\n", err)
		}

		// Process each event
		var offset uint32
		for offset < uint32(n) {
			event := (*syscall.InotifyEvent)(unsafe.Pointer(&buf[offset]))
			// event can also be syscall.IN_DELETE or syscall.IN_ATTRIB
			if event.Mask&syscall.IN_MODIFY == syscall.IN_MODIFY {
				// the file has been modified, so send signal,
				// then wait for the next syscall.Read
				select {
				case <-ctx.Done():
					return nil
				case ch <- struct{}{}:
				}
				break
			}
			offset += syscall.SizeofInotifyEvent + event.Len
		}
	}
}
