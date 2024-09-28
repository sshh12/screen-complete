package main

import (
	"fmt"
	"image"
	"time"

	"bytes"
	"context"
	"encoding/base64"
	"image/png"
	"os"

	"git.tcp.direct/kayos/sendkeys"
	"github.com/kbinani/screenshot"
	"golang.design/x/hotkey"
	"golang.design/x/hotkey/mainthread"

	"github.com/Azure/azure-sdk-for-go/sdk/ai/azopenai"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
)

func main() {
	mainthread.Init(run)
}

func run() {
	// Set up hotkey listener for Ctrl+Shift+S
	hk := hotkey.New([]hotkey.Modifier{hotkey.ModCtrl}, hotkey.KeyQ)
	err := hk.Register()
	if err != nil {
		fmt.Println("Failed to register hotkey:", err)
		return
	}

	fmt.Printf("Hotkey %v is registered\n", hk)

	// Initialize sendkeys
	k, err := sendkeys.NewKBWrapWithOptions(sendkeys.Noisy)
	if err != nil {
		fmt.Println("Failed to initialize sendkeys:", err)
		return
	}

	for {
		<-hk.Keydown()
		<-hk.Keyup()
		time.Sleep(100 * time.Millisecond)
		fmt.Printf("Hotkey %v is pressed\n", hk)

		// Take screenshot
		img, err := screenshot.CaptureDisplay(0)
		if err != nil {
			fmt.Println("Failed to capture screen:", err)
			continue
		}

		systemPrompt := "You are a screenshot based typing assistant. Given a screenshot (including cursor position, highlighted text), write the command/prompt the user is likely trying to convey to you."
		userPrompt := "Here's my current screen and cursor position. Convert this context into a prompt that I am likely trying to convey to you. I will always be asking you to type something (rather than click, open a window, etc). Start with 'Please type...'"
		response := sendToOpenAI(img, systemPrompt, userPrompt)
		fmt.Println("prompt ->", response)

		systemPrompt = "You are a typing assistant. You take a screenshot and a command and type directly onto the screen with your response. Respond only with text that will be typed on the screen. You may not use new lines or special formatting. Keep in mind the formatting the text near the cursor."
		userPrompt = response
		result := sendToOpenAI(img, systemPrompt, userPrompt)
		fmt.Println("result ->", result)

		k.Type(result)
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
