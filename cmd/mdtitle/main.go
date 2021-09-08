package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/knaka/mdpp"
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
	for _, inPath := range os.Args[1:] {
		title := mdpp.GetMarkdownTitle(inPath)
		fmt.Println(title)
	}
}
