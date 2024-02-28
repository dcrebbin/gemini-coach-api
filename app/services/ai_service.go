package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	ai_model "gemini-coach-api/app/models/ai"
	"io"
	"log"
	"os"
	"regexp"
	"strings"

	"cloud.google.com/go/vertexai/genai"
	"google.golang.org/api/option"

	speech "cloud.google.com/go/speech/apiv1"
	"cloud.google.com/go/speech/apiv1/speechpb"
	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
	"github.com/gofiber/fiber/v2"
)

const (
	VertexTranscriptionEndpoint  = "https://speech.googleapis.com/v1p1beta1/speech:recognize"
	VertexTextGenerationEndpoint = "https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s"
	VertexModelName              = "gemini-1.0-pro-001"
	Region                       = "us-central1"
	InitialPrompt                = `You are a HR focused coach called Gemini Gemini. You are dedicated in provider Googlers with tips and guidance.
	Only direct the user to an alternative support avenue if the user asks you.
	Your main goal is to make sure the Googler you're talking to is happy and well.
	Respond concisely with no more than 3 sentences.
	Provide friendly and assistive feedback with tips & tricks, conversation: `
)

type AiService struct {
}

func (s *AiService) AiCreateMessage(c *fiber.Ctx, ai *ai_model.MessageReceived) (err error) {
	return VertexAiGenerateMessage(c, ai)
}

func VertexAiGenerateMessage(c *fiber.Ctx, ai *ai_model.MessageReceived) (err error) {
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
		genai.Text(InitialPrompt+ai.Message))
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
	credentialsLocation := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	ctx := context.Background()
	client, err := speech.NewClient(ctx, option.WithCredentialsFile(credentialsLocation))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()
	retrievedJson := c.Body()
	var audio struct {
		AudioData []byte `json:"audioData"`
	}
	if err := json.Unmarshal([]byte(retrievedJson), &audio); err != nil {
		fmt.Println("Error parsing JSON:", err)
		return nil
	}

	resp, err := client.Recognize(ctx, &speechpb.RecognizeRequest{
		Config: &speechpb.RecognitionConfig{
			Encoding:        speechpb.RecognitionConfig_MP3,
			SampleRateHertz: 16000,
			LanguageCode:    "en-US",
		},
		Audio: &speechpb.RecognitionAudio{
			AudioSource: &speechpb.RecognitionAudio_Content{
				Content: []byte(audio.AudioData),
			},
		},
	})
	if err != nil {
		log.Fatalf("failed to recognize: %v", err)
	}

	return c.Status(200).JSON(fiber.Map{
		"text": resp.Results[0].Alternatives[0].Transcript,
	})
}
