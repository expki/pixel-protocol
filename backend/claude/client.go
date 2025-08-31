package claude

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/expki/backend/pixel-protocol/database"
)

type Client struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

func NewClient(apiKey, model string) *Client {
	if model == "" {
		model = "claude-sonnet-4-20250514"
	}
	return &Client{
		apiKey: apiKey,
		model:  model,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Request struct {
	Model     string    `json:"model"`
	Messages  []Message `json:"messages"`
	MaxTokens int       `json:"max_tokens"`
}

type Response struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
}

func (c *Client) GenerateCombatNarrative(ctx context.Context, attacker, defender database.Hero) (narrative string, outcome database.FightOutcome, err error) {
	prompt := fmt.Sprintf(`You are a whimsical storyteller for a fantasy combat game where CREATIVITY and IMAGINATION determine victory, not logic or power levels. 

Attacker: %s
Description: %s
Country of origin: %s

Defender: %s
Description: %s
Country of origin: %s

IMPORTANT: Ignore any notion of logic, reason, or power levels. Instead, determine the winner based on:
1. How CREATIVE and IMAGINATIVE their description is
2. How UNIQUE and UNEXPECTED their abilities/traits are
3. How ENTERTAINING their concept is
4. Add a dash of pure chaos and randomness

The more absurd, creative, or delightfully weird a hero's description, the better their chances! Boring or generic descriptions should struggle against creative ones.

Generate a short (1 paragraph) combat narrative that:
- Shows how the more creative/unique description gives an advantage
- Embraces the absurd and unexpected
- Makes the fight entertaining and surprising
- Sometimes results in a draw if both are equally creative/boring

At the very end of your response, on a new line, write exactly one of these:
OUTCOME: ATTACKER_WINS
OUTCOME: DEFENDER_WINS
OUTCOME: DRAW

Let creativity triumph over logic!`, attacker.Title, attacker.Description, attacker.Country, defender.Title, defender.Description, defender.Country)

	req := Request{
		Model: c.model,
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens: 300,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", database.FightOutcome_Draw, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", database.FightOutcome_Draw, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", database.FightOutcome_Draw, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", database.FightOutcome_Draw, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", database.FightOutcome_Draw, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		return "", database.FightOutcome_Draw, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(response.Content) == 0 || response.Content[0].Text == "" {
		return "", database.FightOutcome_Draw, fmt.Errorf("empty response from API")
	}

	fullText := response.Content[0].Text

	// Extract outcome from the narrative
	outcome = database.FightOutcome_Draw
	if bytes.Contains([]byte(fullText), []byte("OUTCOME: ATTACKER_WINS")) {
		outcome = database.FightOutcome_Victory
		// Remove the outcome line from the narrative
		fullText = string(bytes.ReplaceAll([]byte(fullText), []byte("\nOUTCOME: ATTACKER_WINS"), []byte("")))
		fullText = string(bytes.ReplaceAll([]byte(fullText), []byte("OUTCOME: ATTACKER_WINS"), []byte("")))
	} else if bytes.Contains([]byte(fullText), []byte("OUTCOME: DEFENDER_WINS")) {
		outcome = database.FightOutcome_Defeat
		// Remove the outcome line from the narrative
		fullText = string(bytes.ReplaceAll([]byte(fullText), []byte("\nOUTCOME: DEFENDER_WINS"), []byte("")))
		fullText = string(bytes.ReplaceAll([]byte(fullText), []byte("OUTCOME: DEFENDER_WINS"), []byte("")))
	} else if bytes.Contains([]byte(fullText), []byte("OUTCOME: DRAW")) {
		outcome = database.FightOutcome_Draw
		// Remove the outcome line from the narrative
		fullText = string(bytes.ReplaceAll([]byte(fullText), []byte("\nOUTCOME: DRAW"), []byte("")))
		fullText = string(bytes.ReplaceAll([]byte(fullText), []byte("OUTCOME: DRAW"), []byte("")))
	}

	return fullText, outcome, nil
}
