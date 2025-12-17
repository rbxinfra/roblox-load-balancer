package daemon

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/golang/glog"
	"github.com/xcdb/syncx"
	"github.rbx.com/roblox/roblox-load-balancer/configuration"
	"github.rbx.com/roblox/roblox-load-balancer/haproxy"
	"github.rbx.com/roblox/roblox-load-balancer/services/types"
)

var (
	gDaemonOnceFlag        sync.Once
	gDaemonCloseSignal     chan struct{} = make(chan struct{}, 1)
	gDaemonCloseSignalWait sync.WaitGroup
	gContextCancelFunc     context.CancelFunc

	gRefreshSignal     chan os.Signal
	gRefreshCancelFunc context.CancelFunc
	gRefreshEvent      *syncx.ManualResetEvent

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

// HandleRemoteRefreshRequest handles any SIGUSR1 commands.
func HandleRemoteRefreshRequest() {
	glog.Infoln("Handling SIGUSR1 requests...")

	gDaemonCloseSignalWait.Add(1)

	for {
		gRefreshSignal = make(chan os.Signal, 1)
		signal.Notify(gRefreshSignal, syscall.SIGUSR1)
		sig := <-gRefreshSignal
		if sig == nil {
			signal.Stop(gRefreshSignal)

			break
		}

		glog.Infoln("Received a SIGUSR1, doing manual configuration reload...")
		gRefreshEvent.Signal()

		signal.Stop(gRefreshSignal)
	}

	gDaemonCloseSignalWait.Done()
}

// Run starts the main Daemon process.
func Run(config *configuration.Config) {
	gDaemonOnceFlag.Do(func() {
		glog.Infoln("Starting daemon thread!")

		gRefreshEvent = syncx.NewManualResetEvent(false)

		gDaemonCloseSignalWait.Add(1)
		ctx, cancel := context.WithCancel(context.Background())

		gContextCancelFunc = cancel

	daemon_loop:
		for {
			select {
			case <-gDaemonCloseSignal:
				break daemon_loop
			default:
				timeoutContext, cancel := context.WithTimeout(context.Background(), *config.RefreshInterval)
				gRefreshCancelFunc = cancel

				services, err := UpdateHAProxyConfigurationFile(ctx, config)
				if err != nil {
					glog.Errorf("Got error when updating HAProxy configuration file: %v", err)

					goto refresh_wait
				}

				if shouldReloadHAProxy(services) {
					glog.Infoln("Reloading HAProxy because of service changes.")

					err = haproxy.ReloadHAProxy(config)
					if err != nil {
						glog.Errorf("Got error when reloading HAProxy: %v", err)
					}
				} else {
					glog.V(100).Infoln("Got service update but no changes detected, skipping HAProxy reload.")
				}

			refresh_wait:

				glog.V(100).Infof("Sleeping for %s...", *config.RefreshInterval)

				gRefreshEvent.WaitContext(timeoutContext)
				gRefreshEvent.Reset()
			}
		}

		glog.Infoln("Exiting daemon thread!")

		gDaemonCloseSignalWait.Done()
	})
}

// Exit signals to the daemon thread to stop working.
func Exit() {
	glog.Infoln("Exit requested, signalling Daemon threads...")

	gDaemonCloseSignal <- struct{}{} // Send a close event to daemon thread
	gRefreshSignal <- nil            // Send a close event to sigusr1 thread

	gRefreshCancelFunc() // Cancel a refresh
	gContextCancelFunc() // Cancel any outgoing HTTP request

	gDaemonCloseSignalWait.Wait() // Wait for thread to exit
}
