# Roblox Load Balancer

Dynamic HAProxy configuration builder for Consul.

## Building

Ensure you have [Go 1.20+](https://go.dev/dl/)

1. Clone the repository via `git`:

    ```txt
    git clone git@github.rbx.com:Roblox/roblox-load-balancer.git
    cd roblox-load-balancer
    ```

2. Build via [make](https://www.gnu.org/software/make/)

    ```txt
    make build-debug
    ```

## Usage

`cd src && go run main.go --help` (use the build binary found in the bin directory if you downloaded a prebuilt or built it yourself)

```txt
Usage: roblox-load-balancer
Build Mode: debug
Commit:  
        [-h|--help]
        [--configuration-file-path[=]] [--dry-run]

  -alsologtostderr
        log to standard error as well as files
  -configuration-file-path string
        The path to the static configuration.
  -dry-run
        Reads from Consul, builds the config, and outputs to the file without starting the Daemon or reloading HAProxy.
  -help
        Print usage.
  -log_backtrace_at value
        when logging hits line file:N, emit a stack trace
  -log_dir string
        If non-empty, write log files in this directory
  -log_link string
        If non-empty, add symbolic links in this directory to the log files
  -logbuflevel int
        Buffer log messages logged at this level or lower (-1 means don't buffer; 0 means buffer INFO only; ...). Has limited applicability on non-prod platforms.
  -logtostderr
        log to standard error instead of files
  -stderrthreshold value
        logs at or above this threshold go to stderr (default 2)
  -v value
        log level for V logs
  -vmodule value
        comma-separated list of pattern=N settings for file-filtered logging
```

# Notice

## Usage of Roblox, or any of its assets.

# ***This project is not affiliated with Roblox Corporation.***

The usage of the name Roblox and any of its assets is purely for the purpose of providing a clear understanding of the project's purpose and functionality. This project is not endorsed by Roblox Corporation, and is not intended to be u
sed for any commercial purposes.

Any code in this project was soley produced with or without the assistance of error traces and/or behaviour analysis of public facing APIs.
