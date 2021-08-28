package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"time"

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
	var shouldPrintHelp bool
	flag.BoolVarP(&shouldPrintHelp, "help", "h", false, "Show Help")

	flag.Parse()
	if shouldPrintHelp {
		flag.Usage()
		os.Exit(0)
	}

	waitForDebugger()

	args := flag.Args()
	for _, inPath := range args {
		var err error
		func() {
			var inFile *os.File
			var err error
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
			err = mdx.Preprocess(inFile, bufOut)
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
			log.Fatal("Failed to preprocess")
		}
	}
}
