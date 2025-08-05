package genai_client

import (
	"context"
	"fitness-trainer/internal/domain"
	"fitness-trainer/internal/logger"
	"fmt"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"github.com/opentracing/opentracing-go"
)

var WorkoutResponseSchema = &genai.Schema{
	Type:     genai.TypeObject,
	Enum:     []string{},
	Required: []string{"reasoning"},
	Properties: map[string]*genai.Schema{
		"exercises": {
			Type: genai.TypeArray,
			Items: &genai.Schema{
				Type:     genai.TypeObject,
				Enum:     []string{},
				Required: []string{"id", "name"},
				Properties: map[string]*genai.Schema{
					"id": {
						Type: genai.TypeString,
					},
					"name": {
						Type: genai.TypeString,
					},
				},
			},
		},
		"reasoning": {
			Type: genai.TypeString,
		},
	},
}

type Client struct {
	client         *genai.Client
	modelName      string
	responseSchema *genai.Schema
}

func New(client *genai.Client, modelName string, responseSchema *genai.Schema) *Client {
	return &Client{client: client, modelName: modelName, responseSchema: responseSchema}
}

func (c *Client) CreateCompletion(ctx context.Context, userID domain.ID, systemPrompt, prompt string) (string, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "genai.CreateCompletion")
	defer span.Finish()

	logger.Debugf("creating completion for user %s", userID)
	logger.Debugf("system prompt: %s", systemPrompt)
	logger.Debugf("user prompt: %s", prompt)

	model := c.client.GenerativeModel(c.modelName)

	model.SetTemperature(1.8)
	model.SetTopK(40)
	model.SetTopP(0.9)
	model.SetMaxOutputTokens(8192)
	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{genai.Text(systemPrompt)},
	}
	if c.responseSchema != nil {
		model.ResponseMIMEType = "application/json"
		model.ResponseSchema = c.responseSchema
	}

	session := model.StartChat()

	resp, err := session.SendMessage(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("failed to send message: %w", err)
	}

	var response strings.Builder
	for _, part := range resp.Candidates[0].Content.Parts {
		if text, ok := part.(genai.Text); ok {
			response.WriteString(string(text))
		}
	}

	logger.Debugf("response: %s", response.String())

	return response.String(), nil
}
