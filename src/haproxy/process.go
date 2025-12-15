package haproxy

import (
	"errors"
	"os"
	"os/exec"
	"syscall"

	"github.com/golang/glog"
	ps "github.com/mitchellh/go-ps"
	"github.rbx.com/roblox/roblox-load-balancer/configuration"
)

var gHAProxyProcess *os.Process

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

// RecoverHAProxyProcess recovers an existing HAProxy process if
// it exists on the system for some reason.
func RecoverHAProxyProcess(config *configuration.Config) error {
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

// ReloadHAProxyConfiguration reloads the HAProxy configuration
// file.
//
// If the config file doesn't exist it exits immediately with no error.
// If the process is running it will send a SIGHUP to it.
// Otherwise it will create the process.
func ReloadHAProxyConfiguration(config *configuration.Config) error {
	glog.Infof("Reloading HAProxy configuration!")

	if _, err := os.Stat(config.OutputFilePath); errors.Is(err, os.ErrNotExist) {
		return nil // Take the case of initial config as not exist.
	}

	if err := validateHAProxyConfiguration(config); err != nil {
		return err
	}

	if haproxyRunning() {
		return gHAProxyProcess.Signal(syscall.SIGHUP)
	}

	glog.Infof("HAProxy was not already running, starting new process...")

	cmd := exec.Command(config.HAProxy.Path)
	cmd.Args = append(cmd.Args, config.HAProxy.Args...)
	cmd.Env = config.HAProxy.Env
	cmd.Dir = config.HAProxy.Dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return err
	}

	gHAProxyProcess = cmd.Process

	return nil
}

// KillHAProxy kills the current HAProxy process.
func KillHAProxy() error {
	glog.Infof("Killing HAProxy process!")

	if gHAProxyProcess == nil {
		return errors.New("HAProxy process is nil, not killing...")
	}

	gHAProxyProcess.Signal(syscall.SIGUSR1)
	state, err := gHAProxyProcess.Wait()
	if err != nil {
		return err
	}

	if !state.Exited() {
		return errors.New("HAProxy process did not exit!")
	}

	gHAProxyProcess = nil

	return nil
}
