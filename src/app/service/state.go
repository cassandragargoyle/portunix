package service

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ProcessState represents the state file (~/.portunix/processes.json)
type ProcessState struct {
	Version   string              `json:"version"`
	Instances map[string]Instance `json:"instances"`
}

// Instance represents a running plugin service instance
type Instance struct {
	PluginName string    `json:"plugin_name"`
	PID        int       `json:"pid"`
	Port       int       `json:"port"`
	Mode       string    `json:"mode"` // shared, exclusive
	Sessions   []Session `json:"sessions"`
	Services   []string  `json:"services"`
	StartedAt  time.Time `json:"started_at"`
}

// Session represents a client session bound to an instance
type Session struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

// InstanceKey returns a unique key for state file map
func InstanceKey(pluginName string, port int) string {
	return fmt.Sprintf("%s:%d", pluginName, port)
}

// StateManager manages the processes.json state file with file locking
type StateManager struct {
	stateFile string
	lockFile  string
	mu        sync.Mutex
}

// NewStateManager creates a new state manager
func NewStateManager() (*StateManager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	portunixDir := filepath.Join(homeDir, ".portunix")
	if err := os.MkdirAll(portunixDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create .portunix directory: %w", err)
	}

	return &StateManager{
		stateFile: filepath.Join(portunixDir, "processes.json"),
		lockFile:  filepath.Join(portunixDir, "processes.lock"),
	}, nil
}

// WithLock executes fn while holding the file lock
func (sm *StateManager) WithLock(fn func(state *ProcessState) error) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Acquire file lock
	lockFd, err := os.OpenFile(sm.lockFile, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("failed to open lock file: %w", err)
	}
	defer lockFd.Close()

	if err := platformFileLock(lockFd.Fd()); err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}
	defer platformFileUnlock(lockFd.Fd())

	// Load state
	state, err := sm.loadState()
	if err != nil {
		return err
	}

	// Execute function
	if err := fn(state); err != nil {
		return err
	}

	// Save state
	return sm.saveState(state)
}

// ReadState reads the current state (with lock, but no write-back)
func (sm *StateManager) ReadState() (*ProcessState, error) {
	var result *ProcessState
	err := sm.WithLock(func(state *ProcessState) error {
		result = state
		return nil
	})
	return result, err
}

func (sm *StateManager) loadState() (*ProcessState, error) {
	data, err := os.ReadFile(sm.stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return &ProcessState{
				Version:   "1.0.0",
				Instances: make(map[string]Instance),
			}, nil
		}
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	var state ProcessState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse state file: %w", err)
	}

	if state.Instances == nil {
		state.Instances = make(map[string]Instance)
	}

	return &state, nil
}

func (sm *StateManager) saveState(state *ProcessState) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	return os.WriteFile(sm.stateFile, data, 0644)
}

// GenerateSessionID creates a short random alphanumeric session ID
func GenerateSessionID() string {
	bytes := make([]byte, 3)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// IsProcessAlive checks if a process with given PID is still running
func IsProcessAlive(pid int) bool {
	return platformIsProcessAlive(pid)
}

// findProcessByPID wraps os.FindProcess for use by platform implementations
func findProcessByPID(pid int) (*os.Process, error) {
	return os.FindProcess(pid)
}
