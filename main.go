package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	. "github.com/dave/jennifer/jen"
	"github.com/fatih/color"
)

type node struct {
	indent int
	text   string

	parent   *node
	children []*node
}

func main() {
	if len(os.Args) < 2 {
		getYourShitTogetherAnd("pass a spec file")
	}
	specFile := os.Args[1]

	if len(os.Args) < 3 {
		getYourShitTogetherAnd("pass a package name")
	}
	packageName := os.Args[2]

	contents, err := ioutil.ReadFile(specFile)
	if err != nil {
		getYourShitTogetherAnd(fmt.Sprintf("get a readable spec file: %s", err))
	}

	specBlueprint := string(contents)
	color.Green("Generating file from:")
	fmt.Println(specBlueprint)
	fmt.Println()

	lines := strings.Split(specBlueprint, "\n")
	lines = lines[0 : len(lines)-1]

	rootNode := &node{indent: -1, text: "node description", children: make([]*node, 0, 10)}
	parentNode := rootNode

	for _, line := range lines {
		indent := countLeadingSpace(line)
		text := getNodeText(line)

		newNode := &node{indent: indent, text: text, children: make([]*node, 0, 10)}

		if indent <= parentNode.indent {
			parentNode = backtrack(parentNode, (parentNode.indent - indent + 1))
		}

		parentNode.children = append(parentNode.children, newNode)
		newNode.parent = parentNode
		parentNode = newNode
	}

	f := generateFile(packageName, rootNode.children[0])

	color.Green("Generated: ")
	fmt.Printf("%#v", f)
	fmt.Println()

	generatedFileName := fmt.Sprintf("%s_test.go", packageName)
	color.Green("Saving file to %s", generatedFileName)

	if err := f.Save(generatedFileName); err != nil {
		getYourShitTogetherAnd(fmt.Sprintf("learn to save a file: %s", err))
	}
}

func backtrack(n *node, steps int) *node {
	var parent *node = n
	for i := 0; i < steps; i++ {
		parent = parent.parent
	}
	return parent
}

func generateFile(packageName string, n *node) *File {
	f := NewFile(fmt.Sprintf("%s_test", packageName))
	f.ImportAlias("github.com/onsi/ginkgo", "g")

	s := Var().Id("_").Op("=")
	s.Add(generateChildren(n))

	f.Add(s)

	return f
}

func generateChildren(n *node) *Statement {
	return Qual("github.com/onsi/ginkgo", getNodeFlavour(n.text)).
		Call(Lit(getNodeDescription(n.text)),
			Func().Params().BlockFunc(func(g *Group) {
				for i, child := range n.children {
					g.Add(generateChildren(child))
					if i < len(n.children)-1 {
						g.Add(Line())
					}
				}
			}))
}

func printGraph(n *node) {
	for _, child := range n.children {
		printGraph(child)
	}
}

func getNodeFlavour(line string) string {
	return strings.Split(strings.TrimLeft(line, " "), " ")[0]
}

func getNodeDescription(line string) string {
	return strings.Trim(strings.Split(line, getNodeFlavour(line))[1], " ")
}

func getNodeText(line string) string {
	return strings.TrimLeft(line, " ")
}

func countLeadingSpace(line string) int {
	i := 0
	for _, runeValue := range line {
		if runeValue == ' ' {
			i++
		} else {
			break
		}
	}
	return i
}

func getYourShitTogetherAnd(message string) {
	fmt.Printf("Get your shit together and %s\n\n", message)
	os.Exit(1)
}
