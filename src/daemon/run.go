package daemon

import (
	"context"
	"sync"

	"github.com/golang/glog"
	"github.rbx.com/roblox/roblox-load-balancer/configuration"
	"github.rbx.com/roblox/roblox-load-balancer/haproxy"
	"github.rbx.com/roblox/roblox-load-balancer/services/types"
)

var (
	gDaemonOnceFlag        sync.Once
	gDaemonCloseSignal     chan struct{} = make(chan struct{}, 1)
	gDaemonCloseSignalWait sync.WaitGroup
	gContextCancelFunc     context.CancelFunc

	gLastServicesList []*types.Service
)

func hashEquals(a, b []*types.Service) bool {
	if len(a) != len(b) {
		return false
	}

	hashA, hashB := 0, 0

	for i := range a {
		hashA += int(a[i].Hash())
		hashB += int(b[i].Hash())
	}

	return hashA == hashB
}

func shouldReloadHAProxy(currentServices []*types.Service) bool {
	if gLastServicesList == nil {
		gLastServicesList = currentServices

		return true
	}

	if hashEquals(gLastServicesList, currentServices) {
		return false
	}

	gLastServicesList = currentServices

	return true
}

// Run starts the main Daemon process.
func Run(config *configuration.Config) {
	gDaemonOnceFlag.Do(func() {
		glog.Infof("Starting daemon thread!")

		gDaemonCloseSignalWait.Add(1)
		ctx, cancel := context.WithCancel(context.Background())

		gContextCancelFunc = cancel

	daemon_loop:
		for {
			select {
			case <-gDaemonCloseSignal:
				break daemon_loop
			default:
				services, err := UpdateHAProxyConfigurationFile(ctx, config)
				if err != nil {
					glog.Errorf("Got error when updating HAProxy configuration file: %v", err)

					continue
				}

				if shouldReloadHAProxy(services) {
					glog.Info("Reloading HAProxy configuration because of service changes.")

					err = haproxy.ReloadHAProxyConfiguration(config)
					if err != nil {
						glog.Errorf("Got error when reloading HAProxy configuration: %v", err)
					}
				} else {
					glog.Warning("Got service update but no changes detected, skipping HAProxy reload.")
				}
			}
		}

		glog.Infof("Exiting daemon thread!")

		gDaemonCloseSignalWait.Done()
	})
}

// Exit signals to the daemon thread to stop working.
func Exit() {
	gContextCancelFunc()
	gDaemonCloseSignal <- struct{}{}
	gDaemonCloseSignalWait.Wait()
}
