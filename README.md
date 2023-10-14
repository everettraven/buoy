# buoy

`buoy` is a declarative TUI dashboard for Kubernetes. You define your dashboard in a JSON file and it will fetch the information from your Kubernetes cluster and build a dashboard for viewing the requested content right in your terminal window.

> [!NOTE]
> This project is in the extremely early stages of development and is a hobby project. Use at your own risk.

[![asciicast](https://asciinema.org/a/1J9xkpgc4Q2E7fCDh0wkDdWKe.svg)](https://asciinema.org/a/1J9xkpgc4Q2E7fCDh0wkDdWKe)

## Quickstart

- Clone the repository:
```
git clone https://github.com/everettraven/buoy.git
```

- `cd` into the cloned repository:
```
cd buoy
```

- Build the `buoy` binary:
```
make build
```

_OR_

```
go build -o buoy main.go
```

- Create a KinD cluster:
```
kind create cluster
```

- Run `buoy` using the `test.json` file:
```
./buoy test.json
```

The `test.json` file contains samples for each of the different panel types that `buoy` currently supports. As this is a hobby project very early in the development cycle there are some limitations and things are bound to not work as expected.

## Controls
- `ctrl+c`, `q` will quit the program and exit the tui
- `tab` will switch the active panel
- up and down arrow keys and mouse scroll will move up and down in the active panel

## Contributing

While this is a hobby project and in the early development stages, I'm more than happy to accept contributions. If you use `buoy` and find some problems or have some ideas for features/improvements, file an issue. If you want to contribute code, feel free to pick up any unassigned issues and create a pull request.

Since this is a hobby project responses to issues and/or pull requests are likely to be slow.
