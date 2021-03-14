package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	_ "github.com/ajstarks/svgo"
	svg "github.com/ajstarks/svgo"
)

func main() {
	g := &MakesenseGraph{
		targets: map[string]*target{},
	}
	scanner := bufio.NewScanner(os.Stdin)
	root := g.GetTarget("<ROOT>")
	g.GraphScan(root, scanner, 0)
	g.dump(dot)
}

func targetNameFromLine(line string) string {
	start := strings.Index(line, "`")
	end := strings.Index(line, "'")
	return line[start+1 : end]
}

func (g *MakesenseGraph) GraphScan(root *target, scanner *bufio.Scanner, level int) {
	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine, indentLevel := findIndentAndTrim(line)
		makefileName := ""
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
			target.mustRemake = true
		} else if strings.HasPrefix(trimmedLine, "Pruning file ") {
			targetName := targetNameFromLine(trimmedLine)
			target := g.GetTarget(targetName)
			root.AddChildren(target)
		} else if (strings.HasPrefix(trimmedLine, "Finished prerequisites of target file ") || strings.HasSuffix(trimmedLine, "was considered already.")) && level+1 >= indentLevel {
			targetName := targetNameFromLine(trimmedLine)
			if targetName != root.name {
				os.Stderr.WriteString(fmt.Sprintf("expected `%s` got `%s`\n", root.name, trimmedLine))
			}
			break
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
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
	SVG
	gexf
)

func (g MakesenseGraph) dump(o OutputType) {
	switch o {
	case list:
		g.dumpList(os.Stdout)
	case dot:
		g.dumpDot(os.Stdout)
	case SVG:
		g.dumpSvg(os.Stdout)
	default:
		return
	}
}

func (g MakesenseGraph) dumpList(w io.Writer) {
	for _, t := range g.targets {
		w.Write([]byte(fmt.Sprintf("%s\n", t.name)))
	}
}

func (g MakesenseGraph) dumpDot(w io.Writer) {
	w.Write([]byte("digraph G {\n"))
	for k, v := range g.targets {
		if k == "<ROOT>" {
			w.Write([]byte(fmt.Sprintf("n%d[shape=point, label=\"root\"];\n", v.id)))
		} else {
			w.Write([]byte(fmt.Sprintf("n%d[label=\"%s\", color=\"%s\"];\n", v.id, v.name, "red")))
		}
	}
	for _, v := range g.targets {
		for _, cv := range v.children {
			w.Write([]byte(fmt.Sprintf("n%d -> n%d ; \n", cv.id, v.id)))
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
	name       string
	children   []*target
	mustRemake bool
}

func (root *target) AddChildren(t *target) {
	root.children = append(root.children, t)
}

type MakesenseGraph struct {
	targets    map[string]*target
	nextUnique int
}

func (m *MakesenseGraph) GetTarget(name string) *target {
	if t, exists := m.targets[name]; exists {
		return t
	}
	m.targets[name] = &target{
		name: name,
		id:   m.nextUnique,
	}
	m.nextUnique++
	return m.targets[name]
}
