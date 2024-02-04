package routes

import (
	handler "gemini-coach-api/app/handlers"
	service "gemini-coach-api/app/services"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

func AiRoutes(api fiber.Router, store *session.Store) {
	aiService := &service.AiService{}
	helperService := &service.HelperService{}
	aiHandler := handler.NewAiHandler(aiService, helperService, store)
	ai := api.Group("/ai")

	ai.Post("/generate-audio", aiHandler.GenerateAudio)
	ai.Post("/chunk", aiHandler.ChunkString)
	ai.Post("/message", aiHandler.ReceiveMessage)
	ai.Post("/speech-to-text", aiHandler.SpeechToText)
}
