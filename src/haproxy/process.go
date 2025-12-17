package haproxy

import (
	"errors"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"

	"github.com/golang/glog"
	ps "github.com/mitchellh/go-ps"
	"github.rbx.com/roblox/roblox-load-balancer/configuration"
)

var (
	gHAProxyProcess *os.Process

	gHAProxyStdoutFile *os.File
	gHAProxyStderrFile *os.File
)

func haproxyRunning() bool {
	if gHAProxyProcess == nil {
		return false
	}

	p, _ := os.FindProcess(gHAProxyProcess.Pid)
	if p == nil || p.Signal(syscall.Signal(0)) != nil {
		return false
	}

	return true
}

func recoverHAProxyProcess(config *configuration.Config) error {
	procs, err := ps.Processes()
	if err != nil {
		return err
	}

	for _, proc := range procs {
		if proc.Executable() == config.HAProxy.Path {
			glog.Infof("Found existing HAProxy process with PID %d", proc.Pid())

			gHAProxyProcess, _ = os.FindProcess(proc.Pid())

			return nil
		}
	}

	return nil
}

// InitializeHAProxy initializes the HAProxy logs
// and tries to recover or start HAProxy.
func InitializeHAProxy(config *configuration.Config) error {
	stdoutDirectory := path.Dir(config.HAProxy.StdoutLogFilePath)
	stderrDirectory := path.Dir(config.HAProxy.StderrLogFilePath)

	err := os.MkdirAll(stdoutDirectory, 0755)
	if err != nil {
		return err
	}

	err = os.MkdirAll(stderrDirectory, 0755)
	if err != nil {
		return err
	}

	gHAProxyStdoutFile, err = os.OpenFile(config.HAProxy.StdoutLogFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}

	gHAProxyStderrFile, err = os.OpenFile(config.HAProxy.StderrLogFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}

	glog.Infof("Initialized HAProxy STDOUT file to %s and STDERR file to %s", config.HAProxy.StdoutLogFilePath, config.HAProxy.StderrLogFilePath)

	err = recoverHAProxyProcess(config)
	if err != nil {
		return err
	}

	err = ReloadHAProxy(config)
	if err != nil {
		return err
	}

	return nil
}

func startHAProxy(config *configuration.Config) {
	glog.Infof("Starting HAProxy as: %s %s (max attempts: %d)", config.HAProxy.Path, strings.Join(config.HAProxy.Args, " "), *config.HAProxy.MaxStartAttempts)

	for i := 0; i < *config.HAProxy.MaxStartAttempts; i++ {
		glog.Infof("Starting HAProxy (attempt %d/%d)...", i+1, *config.HAProxy.MaxStartAttempts)

		cmd := exec.Command(config.HAProxy.Path)
		cmd.Args = append(cmd.Args, config.HAProxy.Args...)
		cmd.Env = config.HAProxy.Env
		cmd.Dir = config.HAProxy.Dir
		cmd.Stdout = gHAProxyStdoutFile
		cmd.Stderr = gHAProxyStderrFile

		if err := cmd.Start(); err != nil {
			glog.Errorf("Got error when starting HAProxy: %v", err)

			continue
		}

		gHAProxyProcess = cmd.Process

		glog.Infof("Started HAProxy with PID %d, waiting for exit...", gHAProxyProcess.Pid)

		state, _ := cmd.Process.Wait() // Don't care about error here as it will most likely just be because it was killed from TeardownHAProxy.

		// Code 0 (normal exit) and 130 (SIGINT) are acceptable.
		if state.ExitCode() != 0 && state.ExitCode() != 130 && state.ExitCode() != -1 {
			glog.Errorf("HAProxy process exited with non-zero exit code %d", state.ExitCode())

			continue
		}

		return
	}

	glog.Errorf("Exceeded maximum HAProxy start attempts (%d), giving up!", config.HAProxy.MaxStartAttempts)
}

// ReloadHAProxy reloads the HAProxy configuration
// file.
//
// If the config file doesn't exist it exits immediately with no error.
// If the process is running it will send a SIGHUP to it.
// Otherwise it will create the process.
func ReloadHAProxy(config *configuration.Config) error {
	glog.Infoln("Reloading HAProxy!")

	if _, err := os.Stat(config.OutputFilePath); errors.Is(err, os.ErrNotExist) {
		glog.Infoln("Output HAProxy configuration file does not exist, not reloading HAProxy!")

		return nil // Take the case of initial config as not exist.
	}

	if err := validateHAProxyConfiguration(config); err != nil {
		return err
	}

	if haproxyRunning() {
		glog.Infof("Sending SIGHUP to HAProxy process %d...", gHAProxyProcess.Pid)

		return gHAProxyProcess.Signal(syscall.SIGHUP)
	}

	go startHAProxy(config)

	return nil
}

// TeardownHAProxy kills the current HAProxy process.
func TeardownHAProxy() {
	glog.Infoln("Killing HAProxy process and closing log files!")

	if gHAProxyProcess != nil {
		gHAProxyProcess.Signal(syscall.SIGUSR1) // Graceful shutdown.
		gHAProxyProcess.Wait()
	}

	gHAProxyStdoutFile.Close()
	gHAProxyStderrFile.Close()
}
