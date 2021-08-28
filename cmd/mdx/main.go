package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/mattn/go-isatty"

	flag "github.com/spf13/pflag"

	"github.com/knaka/mdx"
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
	var outPath string
	flag.StringVarP(&outPath, "outfile", "o", "", "Output outFile")
	var shouldPrintHelp bool
	flag.BoolVarP(&shouldPrintHelp, "help", "h", false, "Show Help")

	flag.Parse()
	if shouldPrintHelp {
		flag.Usage()
		os.Exit(0)
	}

	waitForDebugger()

	var outFile *os.File
	var output io.Writer
	if outPath == "" || outPath == "-" {
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

	args := flag.Args()
	if len(args) == 0 {
		args = append(args, "-")
	}
	for _, inPath := range args {
		var err error
		func() {
			var inFile *os.File
			if inPath == "" || inPath == "-" {
				inFile = os.Stdin
			} else {
				var err error
				inFile, err = os.Open(inPath)
				if err != nil {
					log.Fatal("Failed to open inFile outFile: ", inPath)
				}
				defer func() { _ = inFile.Close() }()
			}
			err = mdx.Preprocess(inFile, output)
		}()
		if err != nil {
			log.Fatal("Failed to preprocess")
		}
	}
}
