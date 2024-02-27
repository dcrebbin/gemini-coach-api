package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	ai_model "gemini-coach-api/app/models/ai"
	"io"
	"os"
	"regexp"
	"strings"

	"cloud.google.com/go/vertexai/genai"
	"google.golang.org/api/option"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
	"github.com/gofiber/fiber/v2"
)

const (
	VertexTranscriptionEndpoint  = "https://speech.googleapis.com/v1p1beta1/speech:recognize"
	VertexTextGenerationEndpoint = "https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s"
	VertexModelName              = "gemini-1.0-pro-001"
	Region                       = "us-central1"
)

type AiService struct {
}

func (s *AiService) AiCreateMessage(c *fiber.Ctx, ai *ai_model.MessageReceived) (err error) {
	return VertexAiGenerateMessage(c, ai, "User")
}

func VertexAiGenerateMessage(c *fiber.Ctx, ai *ai_model.MessageReceived, role string) (err error) {
	ctx := context.Background()
	credentialsLocation := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	projectID := os.Getenv("GOOGLE_PROJECT_ID")
	client, err := genai.NewClient(ctx, projectID, Region, option.WithCredentialsFile(credentialsLocation))
	if err != nil {
		return fmt.Errorf("error creating client: %v", err)
	}
	defer client.Close()

	gemini := client.GenerativeModel(VertexModelName)
	chat := gemini.StartChat()

	r, err := chat.SendMessage(
		ctx,
		genai.Text(role+" "+ai.Message))
	if err != nil {
		return err
	}

	part := r.Candidates[0].Content.Parts[0]
	json, err := json.Marshal(part)
	message := string(json)
	message = strings.Replace(message, "\"", "", -1)

	return c.Status(200).JSON(fiber.Map{
		"message": message,
		"status":  "success",
	})
}

func (s *AiService) VertexAiTextToSpeech(message []byte) (output []byte) {
	ctx := context.Background()
	credentialsLocation := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	client, _ := texttospeech.NewClient(ctx, option.WithCredentialsFile(credentialsLocation))
	resp, err := client.SynthesizeSpeech(ctx, &texttospeechpb.SynthesizeSpeechRequest{
		Input: &texttospeechpb.SynthesisInput{
			InputSource: &texttospeechpb.SynthesisInput_Text{Text: string(message)},
		},
		Voice: &texttospeechpb.VoiceSelectionParams{
			LanguageCode: "en-AU",
			Name:         "en-AU-Neural2-B",
		},
		AudioConfig: &texttospeechpb.AudioConfig{
			AudioEncoding: texttospeechpb.AudioEncoding_MP3,
			SpeakingRate:  1.0,
		},
	})

	if err != nil {
		return nil
	}

	reader := io.NopCloser(bytes.NewReader(resp.AudioContent))
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
