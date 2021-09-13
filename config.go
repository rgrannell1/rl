package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/adrg/xdg"
	"github.com/smallnest/ringbuffer"
	"gopkg.in/yaml.v2"
)

// Read RL configuration from a standard file-path.
func ReadConfig(cfg *ConfigOpts) (*ConfigOpts, error) {
	// read configuration
	cfgConn, err := os.Open(cfg.ConfigPath)
	if err != nil {
		return cfg, err
	}
	defer cfgConn.Close()

	var rlCfg RLConfigFile

	decoder := yaml.NewDecoder(cfgConn)
	err = decoder.Decode(&rlCfg)

	if err != nil {
		return cfg, err
	}

	cfg.Config = rlCfg

	return cfg, nil
}

// Create a configuration file, if it doesn't exist already.
func CreateConfigFile(cfg *ConfigOpts) error {
	// -- create the config file if it doesn't exist
	// -- write to file
	_, err := os.Stat(cfg.ConfigPath)

	if errors.Is(err, os.ErrNotExist) {
		// -- the file does not exist, write yaml to a file
		cfgConn, err := os.OpenFile(cfg.ConfigPath, os.O_RDWR|os.O_CREATE, 0700)
		if err != nil {
			return err
		}
		defer func() {
			cfgConn.Close()
		}()

		enc := yaml.NewEncoder(cfgConn)
		encodeErr := enc.Encode(RLConfigFile{false})

		if encodeErr != nil {
			return encodeErr
		}
	} else {
		return err
	}

	return nil
}

// Create a history file, if it doesn't exist already
func CreateHistoryFile(cfg *ConfigOpts) error {
	histConn, err := os.OpenFile(cfg.HistoryPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0700)
	if err != nil {
		return err
	}

	defer func() {
		histConn.Close()
	}()

	return nil
}

// Initialise RL configuration; create required configuration directories
// and files, and return configuration that's already present
func InitConfig() (*ConfigOpts, error) {
	configPath := filepath.Join(xdg.ConfigHome, "rl.yaml")
	dataDir := filepath.Join(xdg.DataHome, "rl")
	historyPath := filepath.Join(dataDir, "history")

	cfg := ConfigOpts{
		historyPath,
		configPath,
		RLConfigFile{},
	}

	// ensure XDG directories exist
	err := os.MkdirAll(xdg.ConfigHome, 0700)
	if err != nil {
		return &cfg, err
	}

	err = os.MkdirAll(dataDir, 0700)
	if err != nil {
		return &cfg, err
	}

	if cfgErr := CreateConfigFile(&cfg); cfgErr != nil {
		return &cfg, cfgErr
	}

	if histErr := CreateHistoryFile(&cfg); histErr != nil {
		return &cfg, histErr
	}

	return ReadConfig(&cfg)
}

var fileLock = sync.Mutex{}

// Write to file history, if
func HistoryWriter(histChan chan *History, cfg *ConfigOpts) {
	histConn, _ := os.OpenFile(cfg.HistoryPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0700)
	writer := bufio.NewWriter(histConn)

	defer func() {
		fileLock.Lock()
		histConn.Close()
		writer.Flush()
		fileLock.Unlock()
	}()

	startTime := time.Now()

	for {
		hist := <-histChan
		hist.StartTime = startTime
		entry, _ := json.MarshalIndent(hist, "", " ")

		fileLock.Lock()
		writer.WriteString(string(entry) + "\n")
		writer.Flush()
		fileLock.Unlock()
	}
}

func StartHistoryWriter(cfg *ConfigOpts) chan *History {
	// write to RL history, if that's configured
	histChan := make(chan *History)

	if cfg.Config.SaveHistory {
		go HistoryWriter(histChan, cfg)
	}

	return histChan
}

func ReadStdin() (*ringbuffer.RingBuffer, int) {
	stdin := ringbuffer.New(1000 * 10) // 10MB

	piped, pipeErr := StdinPiped()

	if pipeErr != nil {
		fmt.Printf("RL: could not inspect whether sdin was piped in.")
		return stdin, 1
	}

	// read from standard input and redirect to subcommands. Input can be infinite,
	// so manage this read from a goroutine an read into a circular buffer
	if piped {
		go StdinReader(stdin)
	}

	return stdin, 0
}

func ValidateConfig() (*ConfigOpts, int) {
	tty, ttyErr := OpenTTY()
	cfg, cfgErr := InitConfig()

	if cfgErr != nil {
		fmt.Printf("RL: Failed to read configuration: %s\n", cfgErr)
		return cfg, 1
	}

	if ttyErr != nil {
		fmt.Printf("RL: could not open /dev/tty. Are you running rl non-interactively?")
		return cfg, 1
	}
	tty.Close()

	return cfg, 0
}
