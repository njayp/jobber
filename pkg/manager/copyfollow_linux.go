package manager

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"
)

func copyFollow(ctx context.Context, path string, w io.Writer) error {
	// Open the file
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// make watcher
	inotifyFd, err := syscall.InotifyInit()
	if err != nil {
		return fmt.Errorf("initializing inotify: %w\n", err)
	}
	defer syscall.Close(inotifyFd)

	// Add a watch for the dir
	dir := filepath.Dir(path)
	watchFd, err := syscall.InotifyAddWatch(inotifyFd, dir, syscall.IN_MODIFY|syscall.IN_CREATE)
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

	// if exitcode file is present, we can return here
	_, err = os.Stat(filepath.Join(dir, exitCodeFileName))
	if !os.IsNotExist(err) {
		if err != nil {
			return err
		}
		// file exists
		return nil
	}

	// Buffer to receive events
	var buf [4096]byte

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			// Iterate through the buffer and print event details
			n, err := syscall.Read(inotifyFd, buf[:])
			if err != nil {
				log.Fatal(err)
			}

			// Iterate through the buffer and print event details
			for i := 0; i < n; {
				rawEvent := (*syscall.InotifyEvent)(unsafe.Pointer(&buf[i]))
				nameLen := int(rawEvent.Len)
				name := ""
				if nameLen > 0 {
					nameBytes := buf[i+syscall.SizeofInotifyEvent : i+syscall.SizeofInotifyEvent+nameLen]
					name = string(nameBytes[:len(nameBytes)-1])
				}

				switch rawEvent.Mask {
				case syscall.IN_CREATE:
					// process has exited, return
					if name == exitCodeFileName {
						return nil
					}
				case syscall.IN_MODIFY:
					// file was modified
					if name == filepath.Base(path) {
						// read file into writer
						_, err = io.Copy(w, file)
						if err != nil {
							return err
						}
						// we can't break here, we might throw
						// away syscall.IN_CREATE
					}
				default:
					fmt.Printf("Event %x on file %s\n", rawEvent.Mask, name)
				}

				i += syscall.SizeofInotifyEvent + nameLen
			}
		}
	}
}
