# Table

The `table` panel is useful when you want to view a set of data in a structured format. For example, a good use case would be if 
you wanted to view all the `Pod`s in a namespace.

## Example

For brevity, we will show an example of viewing all the `Pod`s in the `default` namespace. In this example, the information we want to know is:
- Name
- Labels
- Phase
- Containers

<!-- tabs:start -->

#### **JSON**
```json
{
    "panels": [
        {
            "name": "All Pods",
            "group": "",
            "version": "v1",
            "kind": "Pod",
            "type": "table",
            "namespace": "default",
            "columns": [
                {
                    "header": "Name",
                    "path": "metadata.name"
                },
                {
                    "header": "Labels",
                    "path": "metadata.labels"
                },
                {
                    "header": "Phase",
                    "path": "status.phase"
                },
                {
                    "header": "Containers",
                    "path": "spec.containers.#.name"
                }
            ]
        }
    ]
}
```

#### **YAML**
```yaml
panels:
  - name: All Pods
    group: ""
    version: v1
    kind: Pod
    type: table
    namespace: default
    columns:
      - header: Name
        path: metadata.name
      - header: Labels
        path: metadata.labels
      - header: Phase
        path: status.phase
      - header: Containers
        path: spec.containers.#.name
```

<!-- tabs:end -->

## Controls

- Up and down arrow keys for selecting rows
- Left and right arrow keys for scrolling rows horizontally
- `v` to toggle viewing the full YAML of the selected resource