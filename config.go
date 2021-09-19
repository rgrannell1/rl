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
	"github.com/docopt/docopt-go"
	"github.com/smallnest/ringbuffer"
	"gopkg.in/yaml.v2"
)

// Read RL configuration from a standard file-path. RL configuration
// will be a YAML file
func ReadConfig(cfg *ConfigOpts) (*ConfigOpts, error) {
	cfgConn, err := os.Open(cfg.ConfigPath)
	if err != nil {
		return cfg, err
	}
	defer func() {
		cfgConn.Close()
	}()

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
		cfgConn, err := os.OpenFile(cfg.ConfigPath, os.O_RDWR|os.O_CREATE, USER_READ_WRITE_OCTAL)
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

// Create a history file, if it doesn't exist already. This may not
// actually be used, depending on user-configuration
func CreateHistoryFile(cfg *ConfigOpts) error {
	histConn, err := os.OpenFile(cfg.HistoryPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, USER_READ_WRITE_OCTAL)
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
	// use XDG specification paths for configuration & data
	configPath := filepath.Join(xdg.ConfigHome, "rl.yaml")
	dataDir := filepath.Join(xdg.DataHome, "rl")
	historyPath := filepath.Join(dataDir, "history")

	cfg := ConfigOpts{
		historyPath,
		configPath,
		RLConfigFile{},
	}

	// ensure XDG directories exist
	for _, dir := range []string{xdg.ConfigHome, dataDir} {
		err := os.MkdirAll(dir, USER_READ_WRITE_OCTAL)
		if err != nil {
			return &cfg, err
		}
	}

	if cfgErr := CreateConfigFile(&cfg); cfgErr != nil {
		return &cfg, cfgErr
	}

	if histErr := CreateHistoryFile(&cfg); histErr != nil {
		return &cfg, histErr
	}

	// Read configuration; if it already exists there might be user configuration here
	return ReadConfig(&cfg)
}

// Write to file history when history events are sent via a channel.
// This will not be used if the user has history disabled
func HistoryWriter(histChan chan *History, cfg *ConfigOpts) {
	var historyLock = sync.Mutex{}
	histConn, _ := os.OpenFile(cfg.HistoryPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, USER_READ_WRITE_OCTAL)
	writer := bufio.NewWriter(histConn)

	defer func() {
		historyLock.Lock()
		histConn.Close()
		writer.Flush()
		historyLock.Unlock()
	}()

	startTime := time.Now()

	for {
		hist := <-histChan
		hist.StartTime = startTime
		entry, _ := json.Marshal(hist)

		historyLock.Lock()
		writer.WriteString(string(entry) + "\n")
		writer.Flush()
		historyLock.Unlock()
	}
}

// Depending on configuration, initialise history writer
func StartHistoryWriter(cfg *ConfigOpts) chan *History {
	// write to RL history, if that's configured
	histChan := make(chan *History)

	if cfg.Config.SaveHistory {
		go HistoryWriter(histChan, cfg)
	}

	return histChan
}

// Read standard-input into a circular buffer; stdin can be infinite, and
// often is when using commands like `journalctl`, we don't want to exhaust all memory
// attempting to store it.
func ReadStdin() (*ringbuffer.RingBuffer, int) {
	stdin := ringbuffer.New(STDIN_BUFFER_SIZE)

	piped, pipeErr := StdinPiped()

	if pipeErr != nil {
		fmt.Printf("RL: could not inspect whether stdin was piped into RL: %v\n", pipeErr)
		return stdin, 1
	}

	// read from standard input and redirect to subcommands. Input can be infinite,
	// so manage this read from a goroutine an read into a circular buffer
	if piped {
		go StdinReader(stdin)
	}

	return stdin, 0
}

// Validate user-configuration before starting RL properly
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

// Read the user's SHELL variable from the environment; this will normally be bash or zsh. If it's present,
// just assume it's accurate, the user would have to lie for it to be set incorrectly most likely
func ReadShell() (string, int) {
	shell := os.Getenv("SHELL")
	if shell == "" {
		fmt.Printf("RL: could not determine user's shell (e.g bash, zsh). Ensure $SHELL is set.")
		return shell, 1
	}

	return shell, 0
}

func RLState(opts *docopt.Opts) (LineChangeState, LineChangeCtx, int) {
	execute, execErr := opts.String("--execute")

	code := AuditCommand(&execute)
	if code != 0 {
		return LineChangeState{}, LineChangeCtx{}, code
	}

	if execErr != nil {
		execute = ""
	}

	inputOnly, inputErr := opts.Bool("--input-only")

	if inputErr != nil {
		fmt.Printf("RL: failed to read --input-only option. %v\n", inputErr)
		os.Exit(1)
	}

	_, rerunErr := opts.Bool("--rerun")

	if rerunErr != nil {
		fmt.Printf("RL: failed to read --rerun option. %v\n", rerunErr)
		os.Exit(1)
	}

	shell, code := ReadShell()
	if code != 0 {
		return LineChangeState{}, LineChangeCtx{}, code
	}

	stdin, code := ReadStdin()
	if code != 0 {
		return LineChangeState{}, LineChangeCtx{}, code
	}

	ctx := LineChangeCtx{
		shell,
		inputOnly,
		&execute,
		os.Environ(),
		stdin,
	}

	linebuffer := LineBuffer{}
	state := LineChangeState{
		lineBuffer: &linebuffer,
		cmd:        nil,
	}

	return state, ctx, 0
}
