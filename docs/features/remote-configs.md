# Using remote dashboard configurations

`buoy` has support for specifying a remote dashboard configuration file. The provided path **must** be a path that returns the raw file contents. For example:
```sh
buoy https://raw.githubusercontent.com/everettraven/buoy/main/test.yaml
```