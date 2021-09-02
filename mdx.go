package mdx

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	mtext "github.com/yuin/goldmark/text"
)

// Write TOC to io.Writer with indent
func writeToc(writer io.Writer, wildcard string, indent string, includerPath string) error {
	paths, err := filepath.Glob(wildcard)
	if err != nil {
		return err
	}
	sort.Strings(paths)
	for _, path := range paths {
		title := getMarkdownTitle(path)
		s := title
		if a, err := filepath.Abs(path); err != nil {
			return err
		} else if a != includerPath {
			s = "[" + title + "](" + path + ")"
		}
		if _, err := fmt.Fprintln(writer, indent+"* "+s); err != nil {
			return err
		}
	}
	return nil
}

func writeFileWithIndent(writer io.Writer, pathForCodeBlock string, indent string) (errReturn error) {
	blockInput, err := os.Open(pathForCodeBlock)
	if err != nil {
		return err
	}
	defer func() {
		if err := blockInput.Close(); err != nil {
			errReturn = err
		}
	}()
	scannerBlockInput := bufio.NewScanner(blockInput)
	for scannerBlockInput.Scan() {
		s := scannerBlockInput.Text()
		if _, err := fmt.Fprintln(writer, indent+s); err != nil {
			return err
		}
	}
	return nil
}

func writeStrBeforeSegmentsStart(writer io.Writer, source []byte,
	position int, segments *mtext.Segments, fix int) (int, error) {
	firstSegment := segments.At(0)
	buf := source[position : firstSegment.Start+fix]
	if _, err := writer.Write(buf); err != nil {
		return position, err
	}
	lastSegment := segments.At(segments.Len() - 1)
	return lastSegment.Stop, nil
}

func writeStrBeforeSegmentsStop(writer io.Writer, source []byte,
	position int, segments *mtext.Segments) (int, error) {
	lastSegment := segments.At(segments.Len() - 1)
	buf := source[position:lastSegment.Stop]
	if _, err := writer.Write(buf); err != nil {
		return position, err
	}
	return lastSegment.Stop, nil
}

const strReBegin = `<!-- *(mdx[_a-zA-Z0-9]*)( ([_a-zA-Z][_a-zA-Z0-9]*)=([^ ]*))? *-->`
const strReEnd = `<!-- /(mdx[_a-zA-Z0-9]*) -->`

func PreprocessWithoutDir(writer io.Writer, reader io.Reader) error {
	_, _, err := Preprocess(writer, reader, "", "")
	return err
}

func Preprocess(writerOut io.Writer, reader io.Reader,
	workDir string, inPath string) (foundMdxDirective bool, changed bool, errReturn error) {
	foundMdxDirective = false
	changed = false
	dirSaved, err := os.Getwd()
	if err != nil {
		return foundMdxDirective, changed, err
	}
	defer func() {
		if err := os.Chdir(dirSaved); err != nil {
			errReturn = err
		}
	}()
	if workDir != "" {
		if err := os.Chdir(workDir); err != nil {
			return foundMdxDirective, changed, err
		}
	}
	var absPath string
	absPath, err = filepath.Abs(filepath.Join(workDir, inPath))
	readBuffer := bytes.NewBuffer(nil)
	if _, err := io.Copy(readBuffer, reader); err != nil {
		return foundMdxDirective, changed, err
	}
	source := readBuffer.Bytes()
	writer := bytes.NewBuffer(nil)
	// Position on source
	position := 0
	// Current location on AST
	var location []*ast.Node
	// Current stack of MDX commands
	var mdxStack []mdxElemMethods
	// RE objects are allocated locally to avoid lock among threads
	var reBegin *regexp.Regexp = nil
	var reEnd *regexp.Regexp = nil
	walker := func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			location = location[:len(location)-1]
			return ast.WalkContinue, nil
		}
		location = append(location, &node)
		switch node.Kind() {
		case ast.KindRawHTML:
			rawHtml, ok := node.(*ast.RawHTML)
			if !ok {
				return ast.WalkStop, errors.New("failed to downcast")
			}
			segments := rawHtml.Segments
			segment := segments.At(0)
			text := string(source[segment.Start:segment.Stop])
			if strings.HasPrefix(text, "<!-- mdx") {
				foundMdxDirective = true
				if reBegin == nil {
					reBegin = regexp.MustCompile(strReBegin)
				}
				match := reBegin.FindStringSubmatch(text)
				if match == nil {
					return ast.WalkStop, NewError("could not match regexp", absPath, source, segment.Start)
				}
				command := match[1]
				baseElem := mdxElem{len(location)}
				if command == "mdxlink" {
					key := match[3]
					if key == "href" {
						value := match[4]
						mdxStack = append(mdxStack, &mdxLinkElem{baseElem, value})
					}
				}
			} else if strings.HasPrefix(text, "<!-- /mdx") {
				if reEnd == nil {
					reEnd = regexp.MustCompile(strReEnd)
				}
				match := reEnd.FindStringSubmatch(text)
				command := match[1]
				if len(mdxStack) == 0 {
					return ast.WalkStop, NewError("unexpected inline closing command", absPath, source, segment.Start)
				}
				if mdxStack[len(mdxStack)-1].Name() != command {
					return ast.WalkStop, NewError("unbalanced closing command", absPath, source, segment.Start)
				}
				if command == "mdxlink" {
					elem, ok := mdxStack[len(mdxStack)-1].(*mdxLinkElem)
					if !ok {
						return ast.WalkStop, NewError("failed to downcast mdxlink", absPath, source, segment.Start)
					}
					mdxStack = mdxStack[:len(mdxStack)-1]
					title := getMarkdownTitle(elem.href)
					modified := "[" + title + "](" + elem.href + ")"
					if _, err := fmt.Fprint(writer, modified); err != nil {
						return ast.WalkStop, err
					}
				}
				position = segment.Start
			}
			position, err = writeStrBeforeSegmentsStop(writer, source, position, segments)
			if err != nil {
				return ast.WalkStop, err
			}
		case ast.KindHTMLBlock:
			htmlBlock, ok := node.(*ast.HTMLBlock)
			if !ok {
				return ast.WalkStop, errors.New("failed to downcast htmlblock")
			}
			if htmlBlock.HTMLBlockType != ast.HTMLBlockType2 {
				break
			}
			segments := node.Lines()
			firstLine := segments.At(0)
			txt := string(source[firstLine.Start:firstLine.Stop])
			if strings.HasPrefix(txt, "<!-- mdx") {
				if reBegin == nil {
					reBegin = regexp.MustCompile(strReBegin)
				}
				match := reBegin.FindStringSubmatch(txt)
				command := match[1]
				mdxElem := mdxElem{len(location)}
				switch command {
				case "mdxcode":
					key := match[3]
					if key == "src" {
						mdxStack = append(mdxStack, &mdxCodeElem{mdxElem, match[4]})
					} else {
						return ast.WalkStop, NewError("attribute \"src\" required", absPath, source, firstLine.Start)
					}
				case "mdxtoc":
					key := match[3]
					if key == "pattern" {
						mdxStack = append(mdxStack, &mdxTocElem{mdxElem, match[4]})
					} else {
						return ast.WalkStop, NewError("attribute \"pattern\" required", absPath, source, firstLine.Start)
					}
				default:
					return ast.WalkStop, NewError("unknown MDX command", absPath, source, firstLine.Start)
				}
			} else if strings.HasPrefix(txt, "<!-- /mdx") {
				if reEnd == nil {
					reEnd = regexp.MustCompile(strReEnd)
				}
				match := reEnd.FindStringSubmatch(txt)
				command := match[1]
				if len(mdxStack) == 0 && command != "mdxcode" {
					return ast.WalkStop, NewError("unexpected block closing command", absPath, source, firstLine.Start)
				}
				switch command {
				case "mdxcode":
					if len(mdxStack) > 0 && mdxStack[len(mdxStack)-1].Name() == command && mdxStack[len(mdxStack)-1].Depth() == len(location) {
						mdxStack = mdxStack[:len(mdxStack)-1]
					}
				case "mdxtoc":
					firstSegment := segments.At(0)
					indent := getIndentBeforeSegment(firstSegment, source)
					elem, ok := mdxStack[len(mdxStack)-1].(*mdxTocElem)
					if !ok {
						return ast.WalkStop, NewError("downcast failed", absPath, source, firstLine.Start)

					}
					mdxStack = mdxStack[:len(mdxStack)-1]
					if err := writeToc(writer, elem.pattern, indent, inPath); err != nil {
						return ast.WalkStop, err
					}
					if elem.Name() != command || elem.Depth() != len(location) {
						return ast.WalkStop, NewError("commands do not match", absPath, source, firstLine.Start)
					}
					position = firstSegment.Start - len(indent)
				default:
					return ast.WalkStop, NewError("unknown closing command", absPath, source, firstLine.Start)
				}
			}
			position, err = writeStrBeforeSegmentsStop(writer, source, position, segments)
			if err != nil {
				return ast.WalkStop, err
			}
		case ast.KindCodeBlock:
			fallthrough
		case ast.KindFencedCodeBlock:
			if len(mdxStack) == 0 || mdxStack[len(mdxStack)-1].Name() != "mdxcode" {
				break
			}
			segments := node.Lines()
			if segments.Len() == 0 {
				return ast.WalkStop, NewError("empty fenced code block", absPath, source, position)
			}
			firstSegment := segments.At(0)
			indent := getIndentBeforeSegment(firstSegment, source)
			position, err = writeStrBeforeSegmentsStart(writer, source, position, segments, -len(indent))
			if err != nil {
				return ast.WalkStop, err
			}
			mdxCodeElem1, ok := mdxStack[len(mdxStack)-1].(*mdxCodeElem)
			if !ok {
				return ast.WalkStop, NewError("downcast failed", absPath, source, firstSegment.Start)
			}
			mdxStack = mdxStack[:len(mdxStack)-1]
			err := writeFileWithIndent(writer, mdxCodeElem1.filepath, indent)
			if err != nil {
				return ast.WalkStop, err
			}
		}
		return ast.WalkContinue, nil
	}
	markdown := goldmark.New(
		goldmark.WithExtensions(
			meta.Meta,
		),
	)
	context := parser.NewContext()
	parserOption := parser.WithContext(context)
	doc := markdown.Parser().Parse(mtext.NewReader(source), parserOption)
	// doc.Dump(source, 0)
	if err := ast.Walk(doc, walker); err != nil {
		return foundMdxDirective, changed, err
	}
	if len(mdxStack) != 0 {
		return foundMdxDirective, changed, errors.New("stack not empty")
	}
	_, err = writer.Write(source[position:])
	if err != nil {
		return foundMdxDirective, changed, err
	}
	dest := writer.Bytes()
	if bytes.Compare(source, dest) != 0 {
		changed = true
	}
	if _, err := io.Copy(writerOut, writer); err != nil {
		return foundMdxDirective, changed, err
	}
	return foundMdxDirective, changed, nil
}

func getIndentBeforeSegment(segment mtext.Segment, source []byte) string {
	indent := ""
	for i := segment.Start - 1; i >= 0; i-- {
		r := rune(source[i])
		if r == ' ' || r == '\t' {
			indent = string(r) + indent
		} else {
			break
		}
	}
	return indent
}
