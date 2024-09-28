package main

import (
	"fmt"
	"image"
	"strings"
	"time"

	"bytes"
	"context"
	"encoding/base64"
	"image/png"
	"os"

	"github.com/go-vgo/robotgo"
	"golang.design/x/hotkey"
	"golang.design/x/hotkey/mainthread"

	"github.com/Azure/azure-sdk-for-go/sdk/ai/azopenai"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
)

func takeScreenshot(x1, y1, x2, y2 int) *image.RGBA {
	// Calculate width and height from mouse coordinates
	w := x2 - x1
	h := y2 - y1

	// Ensure positive width and height
	if w < 0 {
		x1, x2 = x2, x1
		w = -w
	}
	if h < 0 {
		y1, y2 = y2, y1
		h = -h
	}

	fmt.Println("Screenshot area:", x1, y1, w, h)

	bitmap := robotgo.CaptureScreen(x1, y1, w, h)
	defer robotgo.FreeBitmap(bitmap)

	// Convert bitmap to image.RGBA
	img := robotgo.ToRGBA(bitmap)
	return img
}

func main() {
	mainthread.Init(run)
}

func run() {
	s := robotgo.ScaleF()
	fmt.Println("s", s)

	hk := hotkey.New([]hotkey.Modifier{hotkey.ModCtrl}, hotkey.KeyQ)
	err := hk.Register()
	if err != nil {
		fmt.Println("Failed to register hotkey:", err)
		return
	}

	fmt.Printf("Hotkey %v is registered\n", hk)

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

		
		img := takeScreenshot(x1, y1, x2, y2)

		systemPrompt := `You are a screenshot based typing assistant. 
Given a screenshot, type the text that is highlighted by the user's cursor.

Respond in the format (without the quotes):

"<text highlighted> in <application/window> related to <general content in that same window>"
`
		userPrompt := "Here's my current screen and cursor position. Respond in the right format."
		ocr := sendToOpenAI(img, systemPrompt, userPrompt)
		fmt.Println("ocr ->", ocr)

		systemPrompt = `You are a typing assistant that works by responding to prompts within screenshots.
Respond only with text that the user wants your to elaborate or answer.
Do not just repeat the text, instead treat it as a prompt.
You may not use new lines or special formatting.
Keep in mind the formatting the text near the cursor.

Examples:
 - <title> -> respond with the title of the related content on the page
 - <type a poem about bananas> -> respond with a poem about bananas
 - <joke> -> respond with a joke related to the content on the page
`
		userPrompt = `Replace this text with the question or result of the command being asked: ` + ocr
		result := sendToOpenAI(img, systemPrompt, userPrompt)
		fmt.Println("result ->", result)
		
		result = strings.ReplaceAll(result, "\n", " ")
		result = strings.TrimSpace(result)

		robotgo.TypeStr(result)
		
	}
}

func sendToOpenAI(img *image.RGBA, systemPrompt, userPrompt string) string {
	// Convert image to base64
	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	if err != nil {
		return "Error encoding image: " + err.Error()
	}
	base64Img := base64.StdEncoding.EncodeToString(buf.Bytes())

	// Initialize Azure OpenAI client
	key := os.Getenv("OPENAI_API_KEY")
	keyCredential := azcore.NewKeyCredential(key)

	client, err := azopenai.NewClientForOpenAI("https://api.openai.com/v1", keyCredential, nil)
	if err != nil {
		return "Error creating Azure OpenAI client: " + err.Error()
	}

	// Prepare the chat completion request
	content := azopenai.NewChatRequestUserMessageContent([]azopenai.ChatCompletionRequestMessageContentPartClassification{
		&azopenai.ChatCompletionRequestMessageContentPartText{
			Text: to.Ptr(userPrompt),
		},
		&azopenai.ChatCompletionRequestMessageContentPartImage{
			ImageURL: &azopenai.ChatCompletionRequestMessageContentPartImageURL{
				URL: to.Ptr("data:image/png;base64," + base64Img),
			},
		},
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	resp, err := client.GetChatCompletions(ctx, azopenai.ChatCompletionsOptions{
		Messages: []azopenai.ChatRequestMessageClassification{
			&azopenai.ChatRequestSystemMessage{
				Content: to.Ptr(systemPrompt),
			},
			&azopenai.ChatRequestUserMessage{
				Content: content,
			},
		},
		MaxTokens:      to.Ptr[int32](1024),
		DeploymentName: to.Ptr("gpt-4o"),
	}, nil)
	if err != nil {
		return "Error getting chat completions: " + err.Error()
	}

	// Extract and return the response
	if len(resp.Choices) > 0 && resp.Choices[0].Message != nil && resp.Choices[0].Message.Content != nil {
		return *resp.Choices[0].Message.Content
	}

	return ""
}
