package models

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetModelsEndpoint(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"models": []gin.H{
			{"id": "openai-gpt-4"},
			{"id": "openai-gpt-4-turbo"},
			{"id": "openai-gpt-4o"},
			{"id": "openai-gpt-4o-mini"},
			{"id": "openai_o1-o1-preview"},
			{"id": "openai_o1-o1-mini"},
			{"id": "openai_o1-o3-mini"},
			{"id": "anthropic-claude-haiku"},
			{"id": "anthropic-claude-sonnet"},
			{"id": "anthropic-claude-opus"},
			{"id": "perplexity-sonar"},
			{"id": "perplexity-sonar-pro"},
			{"id": "perplexity-sonar-reasoning"},
			{"id": "groq-llama-3.3-70b-versatile"},
			{"id": "groq-llama-3.1-8b-instant"},
			{"id": "groq-llama3-70b-8192"},
			{"id": "together-meta-llama/Meta-Llama-3.1-405B-Instruct-Turbo"},
			{"id": "groq-mixtral-8x7b-32768"},
			{"id": "mistral-open-mistral-nemo"},
			{"id": "mistral-mistral-large-latest"},
			{"id": "mistral-mistral-small-latest"},
			{"id": "mistral-codestral-latest"},
			{"id": "groq-deepseek-r1-distill-llama-70b"},
			{"id": "google-gemini-1.5-flash"},
			{"id": "google-gemini-1.5-pro"},
			{"id": "google-gemini-2.0-flash"},
			{"id": "google-gemini-2.0-flash-thinking"},
			{"id": "together-deepseek-ai/DeepSeek-R1"},
			{"id": "xai-grok-2-latest"},
		},
		"default_models": gin.H{
			"chat":         "openai-gpt-4o-mini",
			"quick_ai":     "openai-gpt-4o-mini",
			"commands":     "openai-gpt-4o-mini",
			"api":          "openai-gpt-4o-mini",
			"emoji_search": "openai-gpt-4o-mini",
			"tools":        "raycast-ray1",
		},
	})
}
