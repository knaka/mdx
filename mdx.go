package mdx

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"unicode"
)

type State int

const (
	OnGround State = iota
	WaitingForCodeBlock
	InsideBacktickFencedCodeBlock
	InsideTildeFencedCodeBlock
	InsideIndentedCodeBlock
	InsideTocBlock
)

var reMdxCommand *regexp.Regexp

// var reMdxInlineCommand *regexp.Regexp
var onceReMdxCommand sync.Once

func isBlankString(s string) bool {
	for _, v := range s {
		if !unicode.IsSpace(v) {
			return false
		}
	}
	return true
}

func takeFirstSpaces(s string) string {
	sRet := ""
	for _, v := range s {
		if !unicode.IsSpace(v) {
			break
		}
		sRet += string(v)
	}
	return sRet
}

const backticks = "```"
const tildes = "~~~"

func replaceAllStringSubMatchFunc(re *regexp.Regexp, str string, repl func([]string) string) string {
	result := ""
	lastIndex := 0

	for _, v := range re.FindAllSubmatchIndex([]byte(str), -1) {
		groups := []string{}
		for i := 0; i < len(v); i += 2 {
			groups = append(groups, str[v[i]:v[i+1]])
		}

		result += str[lastIndex:v[0]] + repl(groups)
		lastIndex = v[1]
	}

	return result + str[lastIndex:]
}

func Preprocess(reader io.Reader, writer io.Writer) error {
	state := OnGround
	pathForCodeBlock := ""
	indent := ""
	wildcard := ""
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		// 3 backticks for the block which is replaced
		switch {
		case (state == WaitingForCodeBlock || state == InsideBacktickFencedCodeBlock) &&
			strings.HasPrefix(strings.TrimSpace(line), backticks):
			switch state {
			case WaitingForCodeBlock:
				state = InsideBacktickFencedCodeBlock
				indent = takeFirstSpaces(line)
			case InsideBacktickFencedCodeBlock:
				err := writeFileWithIndent(writer, pathForCodeBlock, indent)
				if err != nil {
					return err
				}
				state = OnGround
				indent = ""
			}
			if _, err := fmt.Fprintln(writer, line); err != nil {
				return err
			}
		case (state == WaitingForCodeBlock || state == InsideTildeFencedCodeBlock) &&
			strings.HasPrefix(strings.TrimSpace(line), tildes):
			switch state {
			case WaitingForCodeBlock:
				state = InsideTildeFencedCodeBlock
				indent = takeFirstSpaces(line)
			case InsideTildeFencedCodeBlock:
				err := writeFileWithIndent(writer, pathForCodeBlock, indent)
				if err != nil {
					return err
				}
				state = OnGround
				indent = ""
			}
			if _, err := fmt.Fprintln(writer, line); err != nil {
				return err
			}
		// First line of indented block
		case (state == WaitingForCodeBlock) &&
			len(line) > 0 &&
			unicode.IsSpace(rune(line[0])) &&
			!isBlankString(line):
			state = InsideIndentedCodeBlock
			indent = takeFirstSpaces(line)
		case state == InsideIndentedCodeBlock:
			if line == "" || strings.HasPrefix(line, indent) {
				continue
			}
			err := writeFileWithIndent(writer, pathForCodeBlock, indent)
			if err != nil {
				return err
			}
			state = OnGround
			indent = ""
			_, err = fmt.Fprintln(writer, "")
			if err != nil {
				return err
			}
			_, err = fmt.Fprintln(writer, line)
			if err != nil {
				return err
			}
		case state == InsideBacktickFencedCodeBlock:
		case state == InsideTildeFencedCodeBlock:
		case strings.TrimSpace(line) == "<!-- } -->":
			switch state {
			case InsideTocBlock:
				if err := writeToc(writer, wildcard); err != nil {
					return err
				}
			default:
				// Ignore
			}
			_, _ = fmt.Fprintln(writer, line)
			state = OnGround
		case state == InsideTocBlock:
		// Put the other lines as are
		case !strings.Contains(line, "<!--") &&
			!strings.Contains(line, "Mdx"):
			_, _ = fmt.Fprintln(writer, line)
		case state == OnGround:
			// MDX command
			onceReMdxCommand.Do(func() {
				reMdxCommand = regexp.MustCompile(`<!-- (Mdx[_a-zA-Z0-9]*)\(([^)]+)\) *({)? -->`)
				// reMdxInlineCommand = regexp.MustCompile(`(<!-- (Mdx[_a-zA-Z0-9]*)\(([^)]+)\) { -->)([^<]*)(<!-- } -->)`)
			})
			subMatches := reMdxCommand.FindStringSubmatch(line)
			// var done = false
			if len(subMatches) >= 2 {
				command := subMatches[1]
				arg := subMatches[2]
				brace := subMatches[3]
				switch command {
				case "MdxReplaceCode":
					if _, err := os.Stat(arg); err == nil {
						state = WaitingForCodeBlock
						pathForCodeBlock = arg
					}
					// done = true
				case "MdxToc":
					paths, err := filepath.Glob(arg)
					if err != nil {
						return err
					}
					if len(paths) > 0 {
						state = InsideTocBlock
						wildcard = arg
						if brace != "{" {
							return errors.New("brace missing")
						}
					}
					// done = true
				default:
				}
			}
			// if !done {
			// 	if reMdxInlineCommand.FindString(line) != "" {
			// 		var err error
			// 		line = replaceAllStringSubMatchFunc(reMdxInlineCommand, line, func(a []string) string {
			// 			cmd := a[2]
			// 			arg := a[3]
			// 			modified := a[4]
			// 			switch cmd {
			// 			case "MdxLink":
			// 				var title string
			// 				title, err = getTitle(arg)
			// 				modified = "[" + title + "](" + arg + ")"
			// 			default:
			// 			}
			// 			return a[1] + modified + a[5]
			// 		})
			// 		if err != nil {
			// 			return err
			// 		}
			// 	}
			// }
			_, _ = fmt.Fprintln(writer, line)
		default:
			return errors.New("something is wrong")
		}
	}
	if scanner.Err() != nil {
		log.Fatal("Fatal")
	}
	if state == InsideIndentedCodeBlock {
		err := writeFileWithIndent(writer, pathForCodeBlock, indent)
		if err != nil {
			return err
		}
		state = OnGround
		indent = ""
	}
	if state != OnGround {
		return errors.New("not ground")
	}
	return nil
}

func getTitle(path string) (string, error) {
	title := path
	input, err := os.Open(path)
	if err != nil {
		return title, err
	}
	defer func() {
		_ = input.Close()
	}()
	scanner := bufio.NewScanner(input)
	firstSeparator := true
	for scanner.Scan() {
		line := scanner.Text()
		if line == "---" {
			if firstSeparator {
				firstSeparator = false
			} else {
				break
			}
		}
		if strings.HasPrefix(line, "title:") ||
			strings.HasPrefix(line, "Title:") {
			fields := strings.Split(line, ":")
			title = strings.TrimSpace(fields[1])
			if title[0] == '"' && title[len(title)-1] == '"' {
				title = title[1 : len(title)-1]
			}
			break
		}
	}
	return title, nil
}

func writeToc(writer io.Writer, wildcard string) error {
	paths, err := filepath.Glob(wildcard)
	if err != nil {
		return err
	}
	sort.Strings(paths)
	for _, path := range paths {
		title, err := getTitle(path)
		if err != nil {
			title = path
		}
		_, _ = fmt.Fprintln(writer, "* ["+title+"]("+path+")")
	}
	return nil
}

func writeFileWithIndent(writer io.Writer, pathForCodeBlock string, indent string) error {
	blockInput, err := os.Open(pathForCodeBlock)
	if err != nil {
		return err
	}
	defer func() {
		_ = blockInput.Close()
	}()
	scannerBlockInput := bufio.NewScanner(blockInput)
	for scannerBlockInput.Scan() {
		s := scannerBlockInput.Text()
		_, _ = fmt.Fprintln(writer, indent+s)
	}
	return nil
}
