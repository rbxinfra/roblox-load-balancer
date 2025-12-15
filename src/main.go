package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/golang/glog"
	"github.rbx.com/roblox/roblox-load-balancer/configuration"
	"github.rbx.com/roblox/roblox-load-balancer/consul"
	"github.rbx.com/roblox/roblox-load-balancer/daemon"
	"github.rbx.com/roblox/roblox-load-balancer/flags"
	"github.rbx.com/roblox/roblox-load-balancer/haproxy"
)

var applicationName string
var buildMode string
var commitSha string

// Pre-setup, runs before main.
func init() {
	flags.SetupFlags(applicationName, buildMode, commitSha)
}

// Main entrypoint.
func main() {
	if *flags.HelpFlag {
		flag.Usage()

		return
	}

	config, err := configuration.ParseConfiguration()
	if err != nil {
		glog.Error(err)

		os.Exit(1)
	}

	if err := consul.InitializeConsul(config); err != nil {
		glog.Error(err)

		os.Exit(1)
	}

	if *flags.DryRun {
		glog.Infof("Doing dry-run to load initial configuration...")

		_, err := daemon.UpdateHAProxyConfigurationFile(context.Background(), config)
		if err != nil {
			glog.Error(err)

			os.Exit(1)
		}

		return
	}

	if err := haproxy.RecoverHAProxyProcess(config); err != nil {
		glog.Fatal(err)	
		
		os.Exit(1)
	}

	if err := haproxy.ReloadHAProxyConfiguration(config); err != nil { // Only works if config exists already
		glog.Error(err)

		os.Exit(1)
	}
	
	go daemon.Run(config)

	osSignal := make(chan os.Signal, 1)
	signal.Notify(osSignal, syscall.SIGABRT, syscall.SIGINT, syscall.SIGTERM)
	defer func ()  {
		sig := <-osSignal

		daemon.Exit()

		haproxy.KillHAProxy()
		glog.Flush()
		close(osSignal)

		glog.Warningf("Recevied signal %s, exiting", sig)
		os.Exit(0)
	}()
}
