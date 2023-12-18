# Item

The `item` panel is useful when you want to view the full YAML output of a particular resource. Typically this is useful when you want
to watch for any changes to this specific object.

## Example

For this example, we want to keep an eye on a `Deployment` named `foo` in the `default` namespace

<!-- tabs:start -->

#### **JSON**
```json
{
    "panels": [
        {
            "name": "Foo Deployment",
            "group": "apps",
            "version": "v1",
            "kind": "Deployment",
            "type": "item",
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
  - name: Foo Deployment
    group: apps
    version: v1
    kind: Deployment
    type: item
    key:
      namespace: default
      name: foo
```

<!-- tabs:end -->

## Controls

- Up and down arrow keys for navigating the viewport
- Page Up and Page Down for jumping up and down in the viewport
- `v` to toggle viewing the full YAML of the selected resource