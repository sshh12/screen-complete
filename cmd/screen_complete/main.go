package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-vgo/robotgo"
	"github.com/sshh12/screen-complete/pkg/keyboard"
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

		systemPrompt := `You are a screenshot based typing assistant. 
		// ... existing code ...`
		userPrompt := "Here's my current screen and cursor position. Respond in the right format."
		ocr := llm.SendToOpenAI(img, systemPrompt, userPrompt)
		fmt.Println("ocr ->", ocr)

		systemPrompt = `You are a typing assistant that works by responding to prompts within screenshots.
		// ... existing code ...`
		userPrompt = `Replace this text with the question or result of the command being asked: ` + ocr
		result := llm.SendToOpenAI(img, systemPrompt, userPrompt)
		fmt.Println("result ->", result)
		
		result = strings.ReplaceAll(result, "\n", " ")
		result = strings.TrimSpace(result)

		keyboard.TypeString(result)
	}
}