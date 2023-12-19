# Logs

The `logs` panel is great for following the logs of any resource that has a `spec.selector` field that maps
to a `Pod`. It can also be directly pointed to an individual `Pod` to stream that `Pod`s logs.

## Examples

### Getting logs from a `Deployment`

For this example, we are going to fetch some logs for a `Deployment` named `foo` in the `default` namespace

<!-- tabs:start -->

#### **JSON**
```json
{
    "panels": [
        {
            "name": "Foo Logs",
            "group": "apps",
            "version": "v1",
            "kind": "Deployment",
            "type": "logs",
            "key": {
                "namespace": "default",
                "name": "foo"
            } 
        }
    ]
}
```

#### **YAML**
```yaml
panels:
  - name: Foo Logs
    group: apps
    version: v1
    kind: Deployment
    type: logs
    key:
      namespace: default
      name: foo
```

<!-- tabs:end -->

### Logs from a `Pod`

For this example, we are going to fetch some logs for a `Pod` named `foo` in the `default` namespace

<!-- tabs:start -->

#### **JSON**
```json
{
    "panels": [
        {
            "name": "Foo Logs",
            "group": "",
            "version": "v1",
            "kind": "Pod",
            "type": "logs",
            "key": {
                "namespace": "default",
                "name": "foo"
            } 
        }
    ]
}
```

#### **YAML**
```yaml
panels:
  - name: Foo Logs
    group: ""
    version: v1
    kind: Pod
    type: logs
    key:
      namespace: default
      name: foo
```

<!-- tabs:end -->

## Controls

- Up and down arrow keys for navigating the viewport
- Page Up and Page Down for jumping up and down in the viewport
- `/` to enter a search mode. This will open a prompt for inputting a search query. When in search mode:
    - `enter` / `return` executes the search query 
    - `ctrl+s` toggles between fuzzy and strict search. strict search will only match lines strictly containing your search term.
    - `esc` will exit search mode and return to the full logs
