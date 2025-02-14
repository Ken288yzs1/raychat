package chat

import (
	"encoding/json"
	"raychat/auth"
	"strings"
	"time"

	"github.com/samber/lo"
)

type UnTypedOpenAIMessage interface {
	GetRole() string
	GetContent() string
	ToRayChatMessage() RayChatMessage
}

type OpenAIRequest struct {
	Model       string                 `json:"model"`
	Messages    []interface{}          `json:"messages"`
	Stream      bool                   `json:"stream"`
	Temperature float64                `json:"temperature"`
	messages    []UnTypedOpenAIMessage `json:"-"`
}

// func (r OpenAIRequest) ToStrOpenAIRequest() OpenAIRequest[string] {
// 	msgs := make([]OpenAIMessage[string], 0, len(r.Messages))
// 	for _, m := range r.Messages {
// 		tmpMsg := m.ToStrOpenAIMessage()
// 		msgs = append(msgs, tmpMsg)
// 	}

// 	return OpenAIRequest[string]{
// 		Model:       r.Model,
// 		Messages:    msgs,
// 		Stream:      r.Stream,
// 		Temperature: r.Temperature,
// 	}
// }

// func GetStrOpenAIMessage()

func (r OpenAIRequest) ToRayChatRequest(a *auth.RaycastAuth) RayChatRequest {
	messages := make([]RayChatMessage, 0, len(r.Messages))
	for _, m := range r.Messages {
		var (
			tmpMsg UnTypedOpenAIMessage
			err    error
		)
		tmpMsg, err = BuildOpenAIStrMessage(m)
		if err != nil {
			tmpMsg, err = BuildOpenAIPartedMessage(m)
			if err != nil {
				continue
			}
		}
		if tmpMsg.GetRole() == "system" {
			continue
		}

		messages = append(messages, tmpMsg.ToRayChatMessage())
	}
	if r.Temperature == 0 {
		r.Temperature = 1
	}

	model, provider := r.GetRequestModel(a)

	resp := RayChatRequest{
		Debug:             false,
		Locale:            "en-CN",
		Provider:          provider,
		Model:             model,
		Temperature:       r.Temperature,
		SystemInstruction: "markdown",
		Messages:          messages,
	}

	additionalSystemInstructions := r.GetSystemMessage().Content
	if additionalSystemInstructions != "" {
		resp.AdditionalSystemInstructions = additionalSystemInstructions
	}

	return resp
}

func (r OpenAIRequest) GetRequestModel(a *auth.RaycastAuth) (string, string) {
	model := r.Model
	supporedModels := lo.Keys(models)
	for _, m := range a.LoginResp.User.AiChatModels {
		supporedModels = append(supporedModels, m.Model)
	}
	if a.LoginResp.User.EligibleForGpt4 {
		supporedModels = append(supporedModels, "gpt-4")
	}

	if !lo.Contains(supporedModels, r.Model) {
		model = "gpt-3.5-turbo"
	}
	return model, models[model]
}

func (r OpenAIRequest) GetSystemMessage() OpenAIStrMessage {
	additionalSystem := ""
	for _, m := range r.messages {
		if m.GetRole() == "system" {
			additionalSystem = m.GetContent()
		}
	}
	return OpenAIStrMessage{
		Role:    "system",
		Content: additionalSystem,
	}
}

func (r OpenAIRequest) GetNoneSystemMessage() []UnTypedOpenAIMessage {
	msgs := []UnTypedOpenAIMessage{}
	for _, m := range r.messages {
		if m.GetRole() != "system" {
			msgs = append(msgs, m)
		}
	}
	return msgs
}

type RayChatRequest struct {
	Debug                        bool             `json:"debug"`
	Locale                       string           `json:"locale"`
	Messages                     []RayChatMessage `json:"messages"`
	Source                       string           `json:"source"`
	Provider                     string           `json:"provider"`
	Model                        string           `json:"model"`
	Temperature                  float64          `json:"temperature"`
	SystemInstruction            string           `json:"system_instruction"`
	AdditionalSystemInstructions string           `json:"additional_system_instructions,omitempty"`
}

type Content struct {
	Text string `json:"text"`
}

type RayChatMessage struct {
	Content Content `json:"content"`
	Author  string  `json:"author"`
}

func (m RayChatMessage) ToOpenAIMessage() OpenAIStrMessage {
	return OpenAIStrMessage{
		Role:    m.Author,
		Content: m.Content.Text,
	}
}

type RayChatStreamResponse struct {
	Text         string      `json:"text"`
	Reasoning    string      `json:"reasoning"`
	FinishReason *string     `json:"finish_reason"`
	Err          interface{} `json:"error"`
}

func (r RayChatStreamResponse) FromEventString(origin string) RayChatStreamResponse {
	selection := strings.Replace(origin, "data: ", "", 1)
	if len(selection) == 0 {
		return RayChatStreamResponse{}
	}
	err := json.Unmarshal([]byte(selection), &r)
	if err != nil {
		panic(err)
	}
	if strings.Contains(selection, "error") {
		Logger().WithError(err).Errorf("request to raycast error, body: %+v", origin)
		return RayChatStreamResponse{}
	}
	return r
}

func (r RayChatStreamResponse) ToOpenAISteamResponse(model string) OpenAIStreamResponse {

	resp := OpenAIStreamResponse{
		ID:      "chatcmpl-" + generateRandomString(29),
		Object:  "chat.completion.chunk",
		Created: int(time.Now().Unix()),
		Model:   model,
		Choices: []StreamChoices{
			{
				Index:        0,
				FinishReason: r.FinishReason,
			},
		},
	}
	if len(r.Text) != 0 || len(r.Reasoning) != 0 {
		resp.Choices[0].Delta = Delta{
			Role:             "assistant",
			Content:          r.Text,
			ReasoningContent: r.Reasoning,
		}
	}

	return resp
}

type RayChatStreamResponses []RayChatStreamResponse

func (r RayChatStreamResponses) ToOpenAIResponse(model string) OpenAIResponse {
	content := ""
	for _, resp := range r {
		content += resp.Text
	}
	reasoning := ""
	for _, resp := range r {
		reasoning += resp.Reasoning
	}
	return OpenAIResponse{
		ID:      "chatcmpl-" + generateRandomString(29),
		Object:  "chat.completion",
		Created: int(time.Now().Unix()),
		Choices: []Choices{
			{
				Index: 0,
				Message: OpenAIStrMessage{
					Role:             "assistant",
					Content:          content,
					ReasoningContent: reasoning,
				},
				FinishReason: lo.ToPtr("stop"),
			},
		},
		Model: model,
		Usage: Usage{
			PromptTokens:     0,
			CompletionTokens: 0,
			TotalTokens:      0,
		},
	}
}

type OpenAIResponse struct {
	ID      string    `json:"id"`
	Object  string    `json:"object"`
	Created int       `json:"created"`
	Model   string    `json:"model"`
	Choices []Choices `json:"choices"`
	Usage   Usage     `json:"usage"`
}

func (o OpenAIResponse) ToEventString() string {
	bytesRsp, err := json.Marshal(o)
	if err != nil {
		panic(err)
	}
	return "data: " + string(bytesRsp)
}

type ChatMessagePart struct {
	Type      string `json:"type,omitempty"`
	Text      string `json:"text,omitempty"`
	Reasoning string `json:"reasoning,omitempty"`
}

type OpenAIStrMessage struct {
	Role             string `json:"role"`
	Content          string `json:"content"`
	ReasoningContent string `json:"reasoning_content,omitempty"`
}

func (m OpenAIStrMessage) GetContent() string {
	return m.Content
}

func (m OpenAIStrMessage) GetRole() string {
	return m.Role
}

type OpenAIPartedMessage struct {
	Role             string            `json:"role"`
	Content          []ChatMessagePart `json:"content"`
	ReasoningContent string            `json:"reasoning_content,omitempty"`
}

func (m OpenAIPartedMessage) GetContent() string {
	return strings.Join(lo.Map(m.Content, func(part ChatMessagePart, _ int) string { return part.Text }), "\n\n")
}

func (m OpenAIPartedMessage) GetRole() string {
	return m.Role
}

func (m OpenAIPartedMessage) ToRayChatMessage() RayChatMessage {
	return m.ToStrOpenAIMessage().ToRayChatMessage()
}

func (m OpenAIPartedMessage) ToStrOpenAIMessage() OpenAIStrMessage {
	return OpenAIStrMessage{
		Role:             m.GetRole(),
		Content:          m.GetContent(),
		ReasoningContent: m.ReasoningContent,
	}
}

func (m OpenAIStrMessage) ToRayChatMessage() RayChatMessage {
	role := m.Role
	if m.Role == "system" {
		role = "user"
	}

	return RayChatMessage{
		Content: Content{
			Text: m.Content,
		},
		Author: role,
	}
}

type Choices struct {
	Index        int              `json:"index"`
	Message      OpenAIStrMessage `json:"message"`
	FinishReason *string          `json:"finish_reason"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type OpenAIStreamResponse struct {
	ID      string          `json:"id"`
	Object  string          `json:"object"`
	Created int             `json:"created"`
	Model   string          `json:"model"`
	Choices []StreamChoices `json:"choices"`
}

func (o OpenAIStreamResponse) ToEventString() string {
	bytesRsp, err := json.Marshal(o)
	if err != nil {
		panic(err)
	}
	return "data: " + string(bytesRsp)
}

type Delta struct {
	Role             string `json:"role,omitempty"`
	Content          string `json:"content,omitempty"`
	ReasoningContent string `json:"reasoning_content,omitempty"`
}

type StreamChoices struct {
	Index        int     `json:"index"`
	Delta        Delta   `json:"delta"`
	FinishReason *string `json:"finish_reason"`
}

type GetAIInfoResponse struct {
	Models        []ModelInfo `json:"models"`
	DefaultModels struct {
		Chat        string `json:"chat"`
		QuickAi     string `json:"quick_ai"`
		Commands    string `json:"commands"`
		API         string `json:"api"`
		EmojiSearch string `json:"emoji_search"`
	} `json:"default_models"`
}

func (m GetAIInfoResponse) SupporedModels() map[string]string {
	models := map[string]string{}
	for _, model := range m.Models {
		models[model.Model] = model.Provider
	}
	return models
}

type ModelInfo struct {
	ID                     string   `json:"id"`
	Name                   string   `json:"name"`
	Description            string   `json:"description"`
	Status                 any      `json:"status"`
	Features               []string `json:"features"`
	Suggestions            []any    `json:"suggestions"`
	InBetterAiSubscription bool     `json:"in_better_ai_subscription"`
	Model                  string   `json:"model"`
	Provider               string   `json:"provider"`
	ProviderName           string   `json:"provider_name"`
	ProviderBrand          string   `json:"provider_brand"`
	Speed                  int      `json:"speed"`
	Intelligence           float64  `json:"intelligence"`
	RequiresBetterAi       bool     `json:"requires_better_ai"`
	Context                int      `json:"context"`
	Capabilities           struct {
		WebSearch       string `json:"web_search,omitempty"`
		ImageGeneration string `json:"image_generation,omitempty"`
	} `json:"capabilities,omitempty"`
}

func BuildOpenAIStrMessage(origin interface{}) (UnTypedOpenAIMessage, error) {
	raw, err := json.Marshal(origin)
	if err != nil {
		return OpenAIStrMessage{}, err
	}
	var message OpenAIStrMessage
	err = json.Unmarshal(raw, &message)
	if err != nil {
		return OpenAIStrMessage{}, err
	}
	return message, nil
}

func BuildOpenAIPartedMessage(origin interface{}) (UnTypedOpenAIMessage, error) {
	raw, err := json.Marshal(origin)
	if err != nil {
		return OpenAIPartedMessage{}, err
	}
	var message OpenAIPartedMessage
	err = json.Unmarshal(raw, &message)
	if err != nil {
		return OpenAIPartedMessage{}, err
	}
	return message, nil
}
