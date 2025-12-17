package haproxy

import (
	"errors"
	"os"
	"os/exec"

	"github.rbx.com/roblox/roblox-load-balancer/configuration"
)

func validateHAProxyConfiguration(config *configuration.Config) error {
	cmd := exec.Command(config.HAProxy.Path, "-c", "-f", config.OutputFilePath)
	err := cmd.Start()
	cmd.Stderr = os.Stderr
	if err != nil {
		return err
	}

	procInfo, err := cmd.Process.Wait()
	if err != nil {
		return err
	}

	exitCode := procInfo.ExitCode()
	if exitCode == 0 || exitCode == 2 { // 0 for valid config, 2 for valid config but will not run.
		return nil
	}

	return errors.New("invalid HAProxy configuration specified.")
}
