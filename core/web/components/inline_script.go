package components

import (
	"fmt"
	"strings"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

func InlineScript(assetPath string) g.Node {
	fileBytes, err := readAsset(assetPath)
	if err != nil {
		panic(fmt.Errorf("inline script %s: %w", assetPath, err))
	}

	script := wrapScript(string(fileBytes))

	return h.Script(
		h.Type("text/javascript"),
		g.Raw(script),
	)
}

func wrapScript(script string) string {
	script = strings.TrimSpace(script)
	if script == "" {
		return script
	}

	return "(function() {\n" +
		"  function init() {\n" +
		script + "\n" +
		"  }\n\n" +
		"  if (document.readyState === \"loading\") {\n" +
		"    document.addEventListener(\"DOMContentLoaded\", init);\n" +
		"  } else {\n" +
		"    init();\n" +
		"  }\n" +
		"})();\n"
}
