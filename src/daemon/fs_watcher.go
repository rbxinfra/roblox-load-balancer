package daemon

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/golang/glog"
)

// fsWatcher watches multiple files for changes, even if they don't exist yet.
type fsWatcher struct {
	paths    []string
	onChange func(path string)
	interval time.Duration
	done     chan struct{}
	wg       sync.WaitGroup
}

func NewFileWatcher(paths []string, onChange func(path string)) *fsWatcher {
	return &fsWatcher{
		paths:    paths,
		onChange: onChange,
		interval: 2 * time.Second,
		done:     make(chan struct{}),
	}
}

func (fw *fsWatcher) Start() {
	for _, path := range fw.paths {
		fw.wg.Add(1)
		go func(p string) {
			defer fw.wg.Done()
			fw.watchPath(p)
		}(path)
	}
}

func (fw *fsWatcher) Stop() {
	close(fw.done)
	fw.wg.Wait()
}

func (fw *fsWatcher) watchPath(path string) {
	for {
		dir := filepath.Dir(path)
		if err := fw.waitForDir(dir); err != nil {
			return
		}

		if err := fw.waitForFile(path); err != nil {
			return
		}

		if done := fw.watchWithFsnotify(path); done {
			return
		}

		glog.V(100).Infof("file %s removed, re-watching...", path)
	}
}

func (fw *fsWatcher) waitForDir(dir string) error {
	for {
		if _, err := os.Stat(dir); err == nil {
			return nil
		}
		select {
		case <-fw.done:
			return os.ErrClosed
		case <-time.After(fw.interval):
		}
	}
}

func (fw *fsWatcher) waitForFile(path string) error {
	for {
		if _, err := os.Stat(path); err == nil {
			return nil
		}
		glog.V(100).Infof("waiting for file: %s", path)
		select {
		case <-fw.done:
			return os.ErrClosed
		case <-time.After(fw.interval):
		}
	}
}

func (fw *fsWatcher) watchWithFsnotify(path string) bool {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		glog.V(100).Infof("failed to create watcher: %v", err)
		return false
	}
	defer watcher.Close()

	dir := filepath.Dir(path)
	if err := watcher.Add(dir); err != nil {
		glog.V(100).Infof("failed to watch dir %s: %v", dir, err)
		return false
	}

	glog.V(100).Infof("watching file: %s", path)

	for {
		select {
		case <-fw.done:
			return true

		case event, ok := <-watcher.Events:
			if !ok {
				return false
			}

			if filepath.Clean(event.Name) != filepath.Clean(path) {
				continue
			}

			// Only detect write and remove as it assumes it already exists
			// and detecting it being renamed is stupid
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Remove) {
				glog.V(100).Infof("change detected: %s (%s)", path, event.Op)
				fw.onChange(path)
			}


		case err, ok := <-watcher.Errors:
			if !ok {
				return false
			}
			glog.V(100).Infof("watcher error: %v", err)
		}
	}
}
