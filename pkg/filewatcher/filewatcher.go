package filewatcher

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/OSSystems/hulk/log"
	"github.com/fsnotify/fsnotify"
)

// FileWatcher is a file watcher
type FileWatcher struct {
	// Changed notifies when file is created or modified
	Changed chan string

	watcher *fsnotify.Watcher
	files   map[string]bool
	cancel  chan bool
}

// NewFileWatcher initializes a new FileWatcher
func NewFileWatcher() (*FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &FileWatcher{
		watcher: watcher,
		files:   make(map[string]bool),
		cancel:  make(chan bool),
		Changed: make(chan string),
	}, nil
}

// Add starts watching filename
func (fw *FileWatcher) Add(filename string) error {
	path := filename
	parent := filepath.Dir(filename)
	exists := true

	for {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			if path == parent {
				return errors.New("Parent directory does not exist")
			}

			exists = false
			path = filepath.Dir(path)

			continue
		}

		break
	}

	err := fw.watcher.Add(path)
	if err != nil {
		return err
	}

	fw.files[filename] = exists

	return nil
}

// Watch watches for file changes
func (fw *FileWatcher) Watch() {
	go func() {
		for {
			select {
			case <-fw.cancel:
				break
			case event := <-fw.watcher.Events:
				if event.Op == fsnotify.Write || event.Op == fsnotify.Create {
					if _, ok := fw.files[event.Name]; ok {
						fw.Changed <- event.Name
						fw.files[event.Name] = true
					}
				}
			case err := <-fw.watcher.Errors:
				log.Error(err)
			}
		}
	}()

	<-fw.cancel
}
