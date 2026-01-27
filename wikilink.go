package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"

	ctp "github.com/catppuccin/go"
	hashtag "go.abhg.dev/goldmark/hashtag"
	wikilink "go.abhg.dev/goldmark/wikilink"
)

func cleanWikilink(linkText string) string {
	target := strings.ToLower(string(linkText))
	target = strings.TrimSuffix(target, ".md")
	return target
}

func extractWikilinks(md []byte) []string {
	parser := goldmark.New(goldmark.WithExtensions(
		&wikilink.Extender{},
		&hashtag.Extender{Variant: hashtag.ObsidianVariant},
	))

	reader := text.NewReader(md)
	doc := parser.Parser().Parse(reader)

	var names []string
	ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		if link, ok := node.(*wikilink.Node); ok {
			// header link thingy
			if len(link.Target) == 0 {
				return ast.WalkContinue, nil
			}

			names = append(names, cleanWikilink(string(link.Target)))
		}
		if tag, ok := node.(*hashtag.Node); ok {
			names = append(names, "#"+strings.ToLower(string(tag.Tag)))
		}

		return ast.WalkContinue, nil
	})

	return names
}

func pathSuffixes(p string) []string {
	p = filepath.Clean(p)

	var result []string
	for {
		base := filepath.Base(p)
		if len(result) == 0 {
			result = append(result, base)
		} else {
			result = append(result, filepath.Join(base, result[len(result)-1]))
		}

		dir := filepath.Dir(p)
		if dir == "." || dir == p {
			break
		}
		p = dir
	}

	return result
}

func extractVaultGraph(vaultRoot string, flavour ctp.Flavour) []*Node {
	linkmap := make(map[string][]string)

	err := filepath.WalkDir(vaultRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Printf("error while walking: %v\n", err)
			return nil
		}

		if d.IsDir() {
			return nil
		}

		data, readError := os.ReadFile(path)
		if readError != nil {
			fmt.Printf("error while reading markdown: %v\n", readError)
			return nil
		}
		if filepath.Ext(path) != ".md" {
			return nil
		}

		rel, err := filepath.Rel(vaultRoot, path)
		if err != nil {
			panic("Non-relative path inside of vault root??")
		}
		linkmap[cleanWikilink(rel)] = extractWikilinks(data)
		return nil
	})

	if err != nil {
		panic(err)
	}

	graph := []*Node{}
	index := make(map[string]*Node)

	for name, _ := range linkmap {
		node := &Node{
			Name:  filepath.Base(name),
			Pos:   randVec(),
			Color: colorize(flavour.Blue()),
		}

		for _, path := range pathSuffixes(name) {
			if _, ok := index[path]; !ok {
				index[path] = node
			}
		}
		graph = append(graph, node)
	}

	for name, links := range linkmap {
		node, ok := index[name]
		if !ok {
			panic("Could not find node " + name)
		}

		for _, link := range links {
			var target *Node
			// fmt.Printf("%s -> %s\n", name, link)
			if target, ok := index[link]; !ok {
				if strings.HasPrefix(link, "#") {
					target = &Node{
						Name:  filepath.Base(link),
						Pos:   randVec(),
						Color: colorize(flavour.Green()),
					}
				} else {
					target = &Node{
						Name:  filepath.Base(link),
						Pos:   randVec(),
						Color: colorize(flavour.Subtext0()),
					}
				}
				index[link] = target
				graph = append(graph, target)
			}
			target = index[link]

			node.Outgoing = append(node.Outgoing, target)
			target.Incoming = append(target.Incoming, node)
			target.LinkCount++
			node.LinkCount++
		}
	}
	return graph
}
