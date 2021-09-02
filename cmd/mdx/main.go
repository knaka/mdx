package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/knaka/mdx"
	"github.com/mattn/go-isatty"

	flag "github.com/spf13/pflag"
)

func waitForDebugger() {
	if os.Getenv("WAIT_FOR_DEBUGGER") != "" {
		log.Println("PID", os.Getpid())
		for {
			err := exec.Command("sh", "-c", fmt.Sprintf("ps w | grep '\\b[d]lv\\b.*\\battach\\b.*\\b%d\\b'", os.Getpid())).Run()
			time.Sleep(1 * time.Second)
			if err == nil {
				break
			}
		}
	}
}

func main() {
	waitForDebugger()
	var outPath string
	flag.StringVarP(&outPath, "outfile", "o", "", "Output outFile")
	var shouldPrintHelp bool
	flag.BoolVarP(&shouldPrintHelp, "help", "h", false, "Show Help")
	var inPlace bool
	flag.BoolVarP(&inPlace, "in-place", "i", false, "Edit file(s) in place")
	flag.Parse()
	if shouldPrintHelp {
		flag.Usage()
		os.Exit(0)
	}
	if inPlace {
		if outPath != "" {
			_, _ = fmt.Fprintln(os.Stderr, "Do not specify \"outfile\" and \"in-place\" simultaneously")
			os.Exit(1)
		}
	} else {
		if outPath == "" {
			outPath = "-"
		}
	}
	args := flag.Args()
	if inPlace {
		for _, inPath := range args {
			var err error
			func() {
				var inFile *os.File
				inFile, err = os.Open(inPath)
				if err != nil {
					log.Fatal("Failed to open inFile outFile: ", inPath)
				}
				defer func() { _ = inFile.Close() }()
				var outFile *os.File
				outFile, err = ioutil.TempFile("", "mdxi")
				if err != nil {
					return
				}
				defer func() {
					_ = outFile.Close()
					_ = os.Remove(outFile.Name())
				}()
				bufOut := bufio.NewWriter(outFile)
				absPath := ""
				if inPath != "" {
					if absPath, err = filepath.Abs(inPath); err != nil {
						log.Fatal("Error", err.Error())
					}
				}
				var changed bool
				_, changed, err = mdx.Preprocess(bufOut, inFile, filepath.Dir(inPath), absPath)
				if err != nil {
					return
				}
				if !changed {
					return
				}
				err = bufOut.Flush()
				if err != nil {
					return
				}
				if err != nil {
					return
				}
				err = inFile.Close()
				if err != nil {
					return
				}
				err = outFile.Close()
				if err != nil {
					return
				}
				err = os.Rename(outFile.Name(), inPath)
				if err != nil {
					return
				}
			}()
			if err != nil {
				log.Fatalln("Failed to preprocess: ", err.Error())
			}
		}
	} else {
		var outFile *os.File
		var output io.Writer
		if outPath == "-" {
			outFile = os.Stdout
		} else {
			var err error
			outFile, err = os.OpenFile(outPath, os.O_RDWR|os.O_CREATE, 0644)
			if err != nil {
				log.Fatal("Failed to open output outFile: ", outPath)
			}
			defer func() { _ = outFile.Close() }()
		}
		if isatty.IsTerminal(outFile.Fd()) {
			output = outFile
		} else {
			bufOut := bufio.NewWriter(outFile)
			defer func() {
				_ = bufOut.Flush()
			}()
			output = bufOut
		}
		if len(args) == 0 {
			args = append(args, "-")
		}
		for _, inPath := range args {
			var err error
			func() {
				var inFile *os.File
				if inPath == "-" {
					inFile = os.Stdin
				} else {
					var err error
					inFile, err = os.Open(inPath)
					if err != nil {
						log.Fatal("Failed to open inFile outFile: ", inPath)
					}
					defer func() { _ = inFile.Close() }()
				}
				absPath := ""
				if inPath != "" {
					if absPath, err = filepath.Abs(inPath); err != nil {
						log.Fatal("Error", err.Error())
					}
				}
				var workDir string
				if inPath == "-" {
					workDir, err = os.Getwd()
					if err != nil {
						log.Fatal("Failed to get working directory")
					}
				} else {
					workDir = filepath.Dir(inPath)
				}
				_, _, err = mdx.Preprocess(output, inFile, workDir, absPath)
			}()
			if err != nil {
				log.Fatal("Failed to preprocess: ", err.Error())
			}
		}
	}
}
