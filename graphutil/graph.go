package graphutil

import (
	"bytes"

	"github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"
)

var (
	gv = graphviz.New()
)

func ToSvg(graph *cgraph.Graph, layoutType, renderFormat string) ([]byte, error) {
	gv.SetLayout(graphviz.Layout(layoutType))
	var buf bytes.Buffer
	if err := gv.Render(graph, graphviz.Format(renderFormat), &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
