package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-vgo/robotgo"
	"github.com/sshh12/screen-complete/pkg/llm"
	"github.com/sshh12/screen-complete/pkg/screenshot"

	"golang.design/x/hotkey"
	"golang.design/x/hotkey/mainthread"
)

func main() {
	mainthread.Init(run)
}

func run() {
	hk := hotkey.New([]hotkey.Modifier{hotkey.ModCtrl}, hotkey.KeyQ)
	err := hk.Register()
	if err != nil {
		fmt.Println("Failed to register hotkey:", err)
		return
	}

	timeSinceLastEvent := time.Time{}
	x1, y1 := 0, 0
	x2, y2 := 0, 0
	
	fmt.Println("Ready.")
	for {
		<-hk.Keydown()
		x1, y1 = robotgo.Location()
		<-hk.Keyup()
		x2, y2 = robotgo.Location()

		if time.Since(timeSinceLastEvent) < 5*time.Second {
			continue
		}

		timeSinceLastEvent = time.Now()
		img := screenshot.CaptureBounds(x1, y1, x2, y2)

		fmt.Println("Generating...")

		ocr := llm.PromptImage(img, llm.SystemAnalyzeScreenshot, llm.UserAnalyzeScreenshot)
		fmt.Println("> ", ocr)

		result := llm.PromptImage(img, llm.SystemComplete, ocr)
		fmt.Println(">> ", result)
		
		result = strings.ReplaceAll(result, "\n", " ")
		result = strings.TrimSpace(result)
		if strings.HasPrefix(result, "\"") && strings.HasSuffix(result, "\"") {
			result = result[1 : len(result)-1]
		}

		robotgo.TypeStr(result)
	}
}