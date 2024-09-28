package llm

import (
	"fmt"
	"image"
	"time"

	"bytes"
	"context"
	"encoding/base64"
	"image/png"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/ai/azopenai"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
)

func PromptImage(img *image.RGBA, systemPrompt, userPrompt string) string {
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
		fmt.Println("Error getting chat completions:", err.Error())
		return ""
	}

	// Extract and return the response
	if len(resp.Choices) > 0 && resp.Choices[0].Message != nil && resp.Choices[0].Message.Content != nil {
		return *resp.Choices[0].Message.Content
	}

	return ""
}