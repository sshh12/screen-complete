package llm

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/ai/azopenai"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"gopkg.in/yaml.v2"
)

const configFilename = "screen_complete.yml"

type Config struct {
	OpenAIAPIKey string `yaml:"openai_api_key"`
}
func loadConfig() (*Config, error) {
	// Try to read from the current directory
	data, err := os.ReadFile(configFilename)
	if err != nil {
		// If not found, try to read from the home directory
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		data, err = os.ReadFile(filepath.Join(homeDir, configFilename))
		if err != nil {
			return nil, fmt.Errorf("config file not found in current or home directory: %w", err)
		}
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

func getAPIKey() string {
	config, err := loadConfig()
	if err == nil && config.OpenAIAPIKey != "" {
		return config.OpenAIAPIKey
	}
	return os.Getenv("OPENAI_API_KEY")
}

func PromptImage(img *image.RGBA, systemPrompt, userPrompt string) string {
	// Convert image to base64
	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	if err != nil {
		return "Error encoding image: " + err.Error()
	}
	base64Img := base64.StdEncoding.EncodeToString(buf.Bytes())

	// Initialize Azure OpenAI client
	key := getAPIKey()
	if key == "" {
		return "Error: OpenAI API key not found in config file or environment variable"
	}
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