package components

import (
	"fmt"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

func InlineStyle(assetPath string) g.Node {
	fileBytes, err := readAsset(assetPath)
	if err != nil {
		panic(fmt.Errorf("inline style %s: %w", assetPath, err))
	}

	return h.StyleEl(
		h.Type("text/css"),
		g.Raw(string(fileBytes)),
	)
}
