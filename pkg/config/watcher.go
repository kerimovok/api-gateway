package config

import (
	"path/filepath"
	"sync"

	"api-gateway/pkg/utils"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	watcher    *fsnotify.Watcher
	configPath string
	reloadFn   func() error
	mu         sync.Mutex
	done       chan struct{}
}

// New creates a new config watcher
func New(configPath string, reloadFn func() error) (*Watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	w := &Watcher{
		watcher:    fsWatcher,
		configPath: configPath,
		reloadFn:   reloadFn,
		done:       make(chan struct{}),
	}

	return w, nil
}

// Start begins watching the config directory
func (w *Watcher) Start() error {
	// Watch the directory containing config files
	configDir := filepath.Dir(w.configPath)
	if err := w.watcher.Add(configDir); err != nil {
		return err
	}

	go w.watch()
	return nil
}

// Stop stops watching for changes
func (w *Watcher) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()

	close(w.done)
	w.watcher.Close()
}

func (w *Watcher) watch() {
	for {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				utils.LogWarn("Watcher event channel closed")
				return
			}

			// Only handle write and create events for the config file
			if (event.Op&fsnotify.Write == fsnotify.Write ||
				event.Op&fsnotify.Create == fsnotify.Create) &&
				event.Name == w.configPath {

				w.mu.Lock()
				if err := w.reloadFn(); err != nil {
					utils.LogError("Failed to reload config", err)
				} else {
					utils.LogInfo("Config reloaded due to file change")
				}
				w.mu.Unlock()
			}

		case <-w.done:
			utils.LogInfo("Config watcher stopped")
			return

		case err, ok := <-w.watcher.Errors:
			if !ok {
				utils.LogWarn("Watcher error channel closed")
				return
			}
			utils.LogError("Config watcher error", err)
		}
	}
}
