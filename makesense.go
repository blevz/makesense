package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	svg "github.com/ajstarks/svgo"
	"github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"
)

var (
	outputFlag = flag.String("type", "dot", "The type of output for makesense to produce, supported: [`dot`, `list`, `gviz`, `json`]")
	layoutType = flag.String("layout", "dot", "The layout engine to use when creating a gviz image, supported: [`dot`, `circo`, `fdp`, `neato`, `nop`, `nop1`, `nop2`, `osage`, `patchwork`, `sfdp`, `twopi`]")
	renderType = flag.String("render", "svg", "The format to render gviz to, supported: [`dot`, `svg`, `png`, `jpg`]")
)

func main() {
	flag.Parse()
	g := &MakesenseGraph{
		Targets: map[string]*target{},
	}
	scanner := bufio.NewScanner(os.Stdin)
	root := g.GetTarget("<ROOT>")
	g.GraphScan(root, scanner, 0)
	outputType := stringToOutputType[*outputFlag]
	g.dump(outputType, os.Stdout)
}

func targetNameFromLine(line string) string {
	start := strings.Index(line, "`")
	if start == -1 {
		start = strings.Index(line, "'")
		if start == -1 {
			log.Fatalf("Cannot find the start of the target name in line: %s", line)
		}
	}
	end := strings.Index(line[start+1:], "'") + start + 1
	return line[start+1 : end]
}

func (g *MakesenseGraph) GraphScan(root *target, scanner *bufio.Scanner, level int) {
	makefileName := ""
	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine, indentLevel := findIndentAndTrim(line)
		if strings.HasPrefix(trimmedLine, "Considering target file") {
			targetName := targetNameFromLine(trimmedLine)
			if makefileName != "" && targetName == makefileName {
				burnScanner(scanner, makefileName)
			}
			child := g.GetTarget(targetName)
			if level+1 >= indentLevel {
				root.AddChildren(child)
				g.GraphScan(child, scanner, indentLevel+1)
			}
		} else if strings.HasPrefix(trimmedLine, "Must remake target ") {
			targetName := targetNameFromLine(trimmedLine)
			target := g.GetTarget(targetName)
			target.MustRemake = true
			target.Cmds = parseCommand(getCommands(scanner))
		} else if strings.HasPrefix(trimmedLine, "Pruning file ") {
			targetName := targetNameFromLine(trimmedLine)
			target := g.GetTarget(targetName)
			root.AddChildren(target)
		} else if (strings.HasPrefix(trimmedLine, "Finished prerequisites of target file ") || strings.HasSuffix(trimmedLine, "was considered already.")) && level+1 >= indentLevel {
			targetName := targetNameFromLine(trimmedLine)
			if targetName != root.Name {
				os.Stderr.WriteString(fmt.Sprintf("expected `%s` got `%s`\n", root.Name, trimmedLine))
			}
			break
		} else if strings.HasPrefix(trimmedLine, "Reading makefile ") {
			makefileName = targetNameFromLine(trimmedLine)
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func getCommands(scanner *bufio.Scanner) []string {
	toReturn := []string{}
	for scanner.Scan() {
		trimmedLine, _ := findIndentAndTrim(scanner.Text())
		if strings.HasPrefix(trimmedLine, "Successfully remade target file ") {
			return toReturn
		}
		toReturn = append(toReturn, trimmedLine)
	}
	return []string{}
}

func parseCommand(cmds []string) []string {
	toReturn := []string{}
	for _, c := range cmds {
		if !(strings.HasPrefix(c, "Putting child ") || strings.HasPrefix(c, "Removing child ") || strings.HasPrefix(c, "Live child ") || strings.HasPrefix(c, "Reaping winning child ")) {
			toReturn = append(toReturn, c)
		}
	}
	return toReturn
}

func burnScanner(scanner *bufio.Scanner, makefileName string) {
	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine, _ := findIndentAndTrim(line)
		if strings.HasPrefix(trimmedLine, "Finished prerequisites of target file ") || strings.HasSuffix(line, "was considered already.") {
			targetName := targetNameFromLine(trimmedLine)
			if targetName == makefileName {
				return
			}
		}
	}
}

func countLeadingSpaces(line string) int {
	return len(line) - len(strings.TrimLeft(line, " "))
}

func findIndentAndTrim(line string) (string, int) {
	leadingSpaces := countLeadingSpaces(line)
	return line[leadingSpaces:], leadingSpaces
}

type OutputType int

const (
	none OutputType = iota
	list
	dot
	gviz
	SVG
	JSON
	gexf
)

var stringToOutputType = map[string]OutputType{
	"none": none,
	"dot":  dot,
	"gv":   gviz,
	"svg":  SVG,
	"gexf": gexf,
	"json": JSON,
}

func (o OutputType) String() string {
	for k, v := range stringToOutputType {
		if v == o {
			return k
		}
	}
	return ""
}

func (g MakesenseGraph) dump(o OutputType, w io.Writer) {
	switch o {
	case list:
		g.dumpList(w)
	case dot:
		g.dumpDot(w)
	case SVG:
		g.dumpSvg(w)
	case gviz:
		g.dumpGraphViz(w)
	case JSON:
		err := g.dumpJson(w)
		if err != nil {
			log.Fatal(err)
		}
	default:
		return
	}
}

func (m MakesenseGraph) dumpJson(w io.Writer) error {
	return json.NewEncoder(w).Encode(m)
}

func (m MakesenseGraph) dumpGraphViz(w io.Writer) {
	g := graphviz.New()
	graph, err := g.Graph(graphviz.Directed)
	if err != nil {
		log.Fatal(err)
	}
	idToNode := map[string]*cgraph.Node{}
	for k, v := range m.Targets {
		nodeId := fmt.Sprintf("n%d", v.id)
		if k == "<ROOT>" {
			n, err := graph.CreateNode(nodeId)
			if err != nil {
				log.Fatal(err)
			}
			n.SetLabel("root")
			n.SetShape("point")
			idToNode["<ROOT>"] = n
		} else {
			n, err := graph.CreateNode(nodeId)
			if err != nil {
				log.Fatal(err)
			}
			n.SetLabel(v.Name)
			n.SetShape("circle")
			n.SetTooltip(strings.Join(v.Cmds, "\n"))
			if v.MustRemake {
				n.SetColor("red")
			} else {
				n.SetColor("green")
			}
			idToNode[v.Name] = n
		}
	}
	for _, v := range m.Targets {
		for _, cv := range v.Children {
			_, err := graph.CreateEdge("", idToNode[cv], idToNode[v.Name])
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	g.SetLayout(graphviz.Layout(*layoutType))
	renderFormat := graphviz.Format(*renderType)
	var buf bytes.Buffer
	if err := g.Render(graph, renderFormat, &buf); err != nil {
		log.Fatal(err)
	}
	w.Write(buf.Bytes())
}

func (g MakesenseGraph) dumpList(w io.Writer) {
	for _, t := range g.Targets {
		w.Write([]byte(fmt.Sprintf("%s\n", t.Name)))
	}
}

func (g MakesenseGraph) dumpDot(w io.Writer) {
	w.Write([]byte("digraph G {\n"))
	for k, v := range g.Targets {
		color := "green"
		if v.MustRemake {
			color = "red"
		}
		if k == "<ROOT>" {
			w.Write([]byte(fmt.Sprintf("n%d[shape=point, label=\"root\"];\n", v.id)))
		} else {
			w.Write([]byte(fmt.Sprintf("n%d[label=\"%s\", color=\"%s\"];\n", v.id, v.Name, color)))
		}
	}
	for _, v := range g.Targets {
		for _, cv := range v.Children {
			w.Write([]byte(fmt.Sprintf("n%d -> n%d ; \n", g.Targets[cv].id, v.id)))
		}
	}
	w.Write([]byte("}\n"))
}

func (g MakesenseGraph) dumpSvg(w io.Writer) {
	s := svg.New(w)
	s.Start(500, 500)
	s.Circle(200, 200, 100)
	s.End()
}

type target struct {
	id         int
	Name       string   ``
	Children   []string `json:",omitempty"`
	Cmds       []string `json:",omitempty"`
	MustRemake bool     ``
}

func (root *target) AddChildren(t *target) {
	root.Children = append(root.Children, t.Name)
}

type MakesenseGraph struct {
	Targets    map[string]*target
	nextUnique int
}

func (m *MakesenseGraph) GetTarget(name string) *target {
	if t, exists := m.Targets[name]; exists {
		return t
	}
	m.Targets[name] = &target{
		Name: name,
		id:   m.nextUnique,
	}
	m.nextUnique++
	return m.Targets[name]
}
