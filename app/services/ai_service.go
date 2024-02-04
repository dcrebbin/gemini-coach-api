package service

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	ai_model "gemini-coach-api/app/models/ai"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"
)

const (
	VertexTranscriptionEndpoint  = "https://speech.googleapis.com/v1p1beta1/speech:recognize"
	VertexTextGenerationEndpoint = "https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s"
)

type AiService struct {
}

func (s *AiService) AiCreateMessage(c *fiber.Ctx, ai *ai_model.MessageReceived) (err error) {
	return VertexAiGenerateMessage(c, ai, "gemini-pro", "User")
}

func VertexAiGenerateMessage(c *fiber.Ctx, ai *ai_model.MessageReceived, llmModel string, role string) (err error) {
	jsonBody := ai_model.GoogleRequest{
		Contents: []ai_model.GoogleRequestContent{
			{
				Parts: []ai_model.GoogleRequestPart{
					{
						Text: role + " " + ai.Message,
					},
				},
			},
		},
		SafetySettings: []ai_model.GoogleRequestSafety{
			{
				Category:  "HARM_CATEGORY_DANGEROUS_CONTENT",
				Threshold: "BLOCK_ONLY_HIGH",
			},
		},
		GenerationConfig: ai_model.GoogleGenerationConfig{
			Temperature:     1.0,
			TopP:            0.8,
			TopK:            10,
			MaxOutputTokens: 125,
		},
	}
	apiKey := os.Getenv("VERTEX_AI_API_KEY") //using the old gen ai maker method
	url := fmt.Sprintf(VertexTextGenerationEndpoint, llmModel, apiKey)
	agent := fiber.Post(url)
	agent.Set("Content-Type", "application/json")
	agent.JSON(jsonBody)
	statusCode, body, errs := agent.Bytes()
	if len(errs) > 0 {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"errs": errs,
		})
	}
	if statusCode == 400 {
		return c.Status(fiber.StatusUnauthorized).JSON(body)
	}
	if statusCode == 401 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	var googleResponse ai_model.GoogleResponse
	err = json.Unmarshal(body, &googleResponse)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"err": err,
		})
	}

	transformedData := TransformGoogleData(googleResponse)

	return c.Status(statusCode).JSON(fiber.Map{
		"message": transformedData.MessageRetrieved,
		"status":  "success",
	})
}

func TransformGoogleData(responseReceived ai_model.GoogleResponse) ai_model.Response {
	return ai_model.Response{
		MessageRetrieved: responseReceived.Candidates[0].Content.Parts[0].Text,
	}
}

func (s *AiService) VertexAiTextToSpeech(message []byte) (output []byte) {
	url := fmt.Sprintf("https://texttospeech.googleapis.com/v1/text:synthesize")
	agent := fiber.Post(url)
	apiKey := os.Getenv("GCLOUD_API_KEY")

	agent.Set("Authorization", "Bearer "+apiKey)
	agent.Set("Accept", "audio/mpeg")
	agent.Set("Content-Type", "application/json; charset=utf-8")
	agent.Set("x-goog-user-project", "up-it-aps")
	vertexAudioRequest := ai_model.GoogleVertexAiRequest{
		Input: ai_model.GoogleVertexAiAudioRequestInput{
			Text: string(message),
		},
		Voice: ai_model.GoogleVertexAiAudioRequestVoice{
			LanguageCode: "en-AU",
			Name:         "en-AU-Neural2-B",
		},
		AudioConfig: ai_model.GoogleVertexAiAudioRequestAudioConfig{
			AudioEncoding: "MP3",
			SpeakingRate:  1.0,
		},
	}

	agent.JSON(vertexAudioRequest)
	_, body, _ := agent.Bytes()
	vertexResponse := ai_model.GoogleVertexAiAudioResponse{}
	err := json.Unmarshal(body, &vertexResponse)
	if err != nil {
		return nil
	}
	decodedBytes, err := base64.StdEncoding.DecodeString(vertexResponse.AudioContent)
	if err != nil {
		fmt.Println("Error decoding Base64 string:", err)
		return
	}
	reader := io.NopCloser(bytes.NewReader(decodedBytes))
	byteArray, _ := io.ReadAll(reader)
	output = byteArray

	return output
}

func (s *AiService) Chunking(input string) (output [][]byte) {
	r := regexp.MustCompile(`[!?.,]`)
	result := r.Split(input, -1)

	for _, v := range result {
		trimmed := strings.TrimSpace(v)
		if trimmed != "" {
			fmt.Println(trimmed)
			output = append(output, []byte(trimmed))
		}
	}

	return output
}

func (s *AiService) VertexAiSpeechToText(c *fiber.Ctx) (err error) {
	url := fmt.Sprintf(VertexTranscriptionEndpoint)
	agent := fiber.Post(url)
	apiKey := os.Getenv("GCLOUD_API_KEY")
	agent.Set("Authorization", "Bearer "+apiKey)
	agent.Set("Content-Type", "application/json; charset=utf-8")
	agent.Set("x-goog-user-project", "up-it-aps") //replace with your project id

	retrievedJson := c.Body()
	var audio struct {
		AudioData []byte `json:"audioData"`
	}

	if err := json.Unmarshal([]byte(retrievedJson), &audio); err != nil {
		fmt.Println("Error parsing JSON:", err)
		return nil
	}

	encodedString := base64.StdEncoding.EncodeToString(audio.AudioData)
	if err != nil {
		fmt.Println("Error decoding Base64 string:", err)
		return
	}

	jsonBody := ai_model.GoogleVertexAiSpeechToTextRequest{
		Config: ai_model.GoogleVertexAiSpeechToTextRequestConfig{
			LanguageCode:          "en-AU",
			EnableWordTimeOffsets: true,
			EnableWordConfidence:  true,
			Model:                 "default",
			Encoding:              "MP3",
			SampleRateHertz:       24000,
			AudioChannelCount:     1,
		},
		Audio: ai_model.GoogleVertexAiSpeechToTextAudio{
			Content: encodedString,
		},
	}
	response := agent.JSON(jsonBody)
	_, body, errs := response.Bytes()
	if len(errs) > 0 {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"errs": errs,
		})
	}

	var vertexResponse ai_model.GoogleVertexAiSpeechToTextResponse
	err = json.Unmarshal(body, &vertexResponse)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"err": err,
		})
	}
	return c.Status(200).JSON(fiber.Map{
		"text": vertexResponse.VertexAiSpeechToTextResponseResults[0].Alternatives[0].Transcript,
	})
}
