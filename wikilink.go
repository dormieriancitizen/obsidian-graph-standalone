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

func extractVaultGraph(vaultRoot string) map[string][]string {
	graph := make(map[string][]string)

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

		graph[cleanWikilink(filepath.Base(path))] = extractWikilinks(data)
		return nil
	})

	if err != nil {
		panic(err)
	}
	return graph
}
