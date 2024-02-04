package handler

import (
	"bufio"
	ai_model "gemini-coach-api/app/models/ai"
	service "gemini-coach-api/app/services"
	"log"

	"github.com/gofiber/fiber/v2/middleware/session"

	"github.com/gofiber/fiber/v2"
)

type AiHandler struct {
	aiService     *service.AiService
	helperService *service.HelperService
	store         *session.Store
}

func NewAiHandler(aiService *service.AiService, helperService *service.HelperService, store *session.Store) *AiHandler {
	return &AiHandler{aiService: aiService, helperService: helperService, store: store}
}

// Placeholder
// @Param request body main.MyHandler.request true "query params"
// @Success 200 {object} main.MyHandler.response
// @Router /test [post]
func (h *AiHandler) ReceiveMessage(c *fiber.Ctx) error {
	log.Println("ReceiveMessage")
	message := new(ai_model.MessageReceived)
	if err := c.BodyParser(message); err != nil {
		return c.Status(400).SendString(err.Error())
	}
	return h.aiService.AiCreateMessage(c, message)

}

func (h *AiHandler) ChunkString(c *fiber.Ctx) error {
	log.Println("ChunkString")
	query := c.Query("text")
	chunkedString := h.aiService.Chunking(query)
	return h.helperService.ChunkData(c, chunkedString)
}

func (h *AiHandler) GenerateChunkedAudio(ctx *fiber.Ctx) (err error) {
	log.Println("GenerateChunkedAudio")
	message := new(ai_model.MessageReceived)
	if err := ctx.BodyParser(message); err != nil {
		log.Println(err)
		return ctx.Status(400).SendString(err.Error())
	}
	chunkedMessage := h.aiService.Chunking(message.Message)
	ctx.Set("Transfer-Encoding", "chunked")

	ctx.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		doneCh := make(chan bool)

		for i := 0; i < len(chunkedMessage); i++ {
			go func(index int) {
				var audio []byte = h.aiService.VertexAiTextToSpeech(chunkedMessage[index])
				_, err := w.Write(audio)
				if err != nil {
					doneCh <- false
					log.Fatal(err)
					return
				}
				err = w.Flush()
				log.Println("Sending chunk")
				if err != nil {
					print(err)
					doneCh <- false
					return
				}
				doneCh <- true
			}(i)

			if !<-doneCh {
				return
			}
		}
	})
	return nil
}

func (h *AiHandler) SpeechToText(c *fiber.Ctx) error {
	log.Println("SpeechToText")
	return h.aiService.VertexAiCreateTranscription(c)
}
