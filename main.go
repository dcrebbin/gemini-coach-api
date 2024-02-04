package main

import (
	_ "gemini-coach-api/docs"
	"gemini-coach-api/pkg/routes"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/swagger"
	"github.com/joho/godotenv"
)

// Start of the sample swagger config
// @title			Fiber Example API
// @version		1.0
// @description	This is a sample swagger for Fiber
// @termsOfService	http://swagger.io/terms/
// @contact.name	API Support
// @contact.email	fiber@swagger.io
// @license.name	Apache 2.0
// @license.url	http://www.apache.org/licenses/LICENSE-2.0.html
// @host			localhost:8080
// @BasePath		/
func main() {
	app := fiber.New(fiber.Config{DisablePreParseMultipartForm: true, StreamRequestBody: true, PassLocalsToViews: true})
	app.Get("/swagger/*", swagger.HandlerDefault) // default

	config := session.Config{
		Expiration:        24 * time.Hour,
		KeyLookup:         "cookie:session",
		CookieSessionOnly: true,
		CookieSecure:      true,
	}
	store := session.New(config)
	dotEnvError := godotenv.Load()
	if dotEnvError != nil {
		println(dotEnvError)
	}

	app.Use(cors.New(cors.Config{
		AllowCredentials: true,
		AllowOrigins:     os.Getenv("ALLOWED_ORIGINS"),
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-API-KEY",
	}))

	assignRoutes(app, store)

	appListeningError := app.Listen("0.0.0.0:8080")
	if appListeningError != nil {
		return
	}
}

func assignRoutes(app *fiber.App, store *session.Store) {
	app.Get("/", healthCheck)
	api := app.Group("/api")
	routes.AiRoutes(api, store)
}

// HealthCheck godoc
//
//	@Summary		Show the status of server, pog.
//	@Description	get the status of server.
//	@Tags			root
//	@Accept			*/*
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}
//	@Router			/ [get]
func healthCheck(c *fiber.Ctx) error {
	res := map[string]interface{}{
		"data": "Server is up and running",
	}

	if err := c.JSON(res); err != nil {
		return err
	}

	return nil
}

func logout(c *fiber.Ctx, store *session.Store) error {
	sess, _ := store.Get(c)
	err := sess.Destroy()
	if err != nil {
		return err
	}
	return c.Next()
}
