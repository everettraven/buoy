# Customizing the theme of your dashboards

`buoy` supports customizing the theme of your dashboard via a theme configuration file. There are two ways this can be done:
1. Specifying the path to the file via the `--theme` flag. Ex: `buoy dash.json --theme path/to/theme.json`
2. Creating a theme file in the default theme configuration file path (`~/.config/buoy/themes/default.json`)

Currently, the theme configuration files must be JSON and looks like:
```json
{
    "tabColor": {
        "light": "63",
        "dark": "117"
    },
    "selectedRowHighlightColor":  {
        "light": "63",
        "dark": "117"
    },
    "logSearchHighlightColor":  {
        "light": "63",
        "dark": "117"
    },
    "syntaxHighlightDarkColor": "nord",
    "syntaxHighlightLightColor": "monokailight"
}
```

?> The example above is a representation of the default theme that is used when no theme configurations are found/specified

- `tabColor` is the adaptive color scheme used when rendering each tab and the borders of the dashboard. `buoy` renders adaptively based on the color of the terminal background so the color code specified in `light` will be the color that is used when rendering on a light terminal background and `dark` will be used when rendering on a dark terminal background. `buoy` uses https://github.com/charmbracelet/lipgloss for styling and thus respects the same color values. The supported values for colors can be found at https://github.com/charmbracelet/lipgloss?tab=readme-ov-file#colors
- `selectedRowHighlightColor` is the adaptive color scheme used to highlight the currently selected row on a `table` panel.
- `logSearchHighlightColor` is the adaptive color scheme used to highlight search results when searching in a `log` panel.
- `syntaxHighlightDarkColor` is the color theme used for syntax highlighting against a dark terminal background. `buoy` uses https://github.com/alecthomas/chroma for syntax highlighting and thus the same color themes. The available color themes can be found at https://github.com/alecthomas/chroma/tree/master/styles
- `syntaxHighlightLightColor` is the color theme used for syntax highlighting against a light terminal background.
