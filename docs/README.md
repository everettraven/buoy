# buoy

`buoy` is a declarative TUI dashboard for Kubernetes. You define your dashboard in a JSON or YAML file and it will fetch the information from your Kubernetes cluster and build a dashboard for viewing the requested content right in your terminal window.

!> This project is in the extremely early stages of development and is a hobby project. Use at your own risk 

[![asciicast](https://asciinema.org/a/625808.svg)](https://asciinema.org/a/625808)

## Motivation

I created `buoy` because I do a lot of work on Kubernetes controllers. When I am making changes, I often find myself typing out a bunch of the same `kubectl ...` commands and switching between them. 
Some of those commands are blocking (i.e `kubectl get logs -f ...`) and to keep them running while running other commands required opening a new terminal window and more typing.
Since I was running pretty repetitive commands I thought there had to be a better solution. I looked through existing CLI tooling around this space, but none had a simple interface that followed the pattern of
"define what you want to see and I'll show it to you". Thus `buoy` was created to fill this gap (and save me some time while delaying the inevitable arthritis).

## Quickstart

Install `buoy` by downloading one of the binaries from the [releases](https://github.com/everettraven/buoy/releases) or by running:
```sh
go install github.com/everettraven/buoy@latest
```

Load a dashboard with:
```sh
buoy <dashboard config file path>
```

## General Controls
- `ctrl+c`, `q` will quit the program and exit the tui
- `tab` will switch the active tab to the one to the right of the currently active tab
- `shift+tab` will switch the active tab to the one to the left of the currently active tab
- `ctrl+h` will open a more detailed help menu

## Contributing

While this is a hobby project and in the early development stages, I'm more than happy to accept contributions. If you use `buoy` and find some problems or have some ideas for features/improvements, file an issue. If you want to contribute code, feel free to pick up any unassigned issues and create a pull request.

Since this is a hobby project responses to issues and/or pull requests are likely to be slow.
