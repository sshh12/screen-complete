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

		// Send screenshot to OpenAI API
		response := strings.TrimSpace(sendToOpenAI(img))

		fmt.Println("resp ->", response)

		k.Type(response)
	}
}

func sendToOpenAI(img *image.RGBA) string {
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
			Text: to.Ptr("Here's my current screen and cursor position. Provide text that I should type next or to fill my selection. Respond only with text that will be typed on the screen."),
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
				Content: to.Ptr("You are a screenshot based magic text completition API. You take a screenshot and intelligently predict what the user is trying to type. If the user has highlighted text that includes '<>'s or '[]'s like <insert title here>, you should interpret that text as a prompt for you answer (generating a title for the text) rather than complete."),
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
