package app

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/JonathanGunawan30/portal-news-backend.git/config"
	"github.com/JonathanGunawan30/portal-news-backend.git/internal/adapter/cloudflare"
	"github.com/JonathanGunawan30/portal-news-backend.git/internal/adapter/handler"
	"github.com/JonathanGunawan30/portal-news-backend.git/internal/adapter/handler/response"
	"github.com/JonathanGunawan30/portal-news-backend.git/internal/adapter/repository"
	"github.com/JonathanGunawan30/portal-news-backend.git/internal/core/service"
	"github.com/JonathanGunawan30/portal-news-backend.git/lib/auth"
	"github.com/JonathanGunawan30/portal-news-backend.git/lib/middleware"
	"github.com/JonathanGunawan30/portal-news-backend.git/lib/pagination"
	"github.com/gofiber/contrib/swagger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/rs/zerolog/log"
)

func RunServer() {
	cfg := config.NewConfig()
	if strings.TrimSpace(cfg.App.JwtSecretKey) == "" {
		log.Fatal().Msg("JWT_SECRET_KEY must be configured")
		return
	}
	if len(cfg.App.JwtSecretKey) < 32 {
		log.Warn().Msg("JWT_SECRET_KEY is shorter than 32 bytes; use a stronger secret key")
	}

	db, err := cfg.ConnectionPostgres()
	if err != nil {
		log.Fatal().Msgf("Error connection to database: %v", err)
		return
	}

	r2Adapter := cloudflare.NewCloudflareR2Adapter(cfg)

	jwtAuth := auth.NewJwt(cfg)
	middlewareAuth := middleware.NewMiddleware(jwtAuth)
	_ = pagination.NewPagination()

	// repository
	authRepository := repository.NewAuthRepository(db.DB)
	categoryRepository := repository.NewCategoryRepository(db.DB)
	contentRepository := repository.NewContentRepository(db.DB)
	userRepository := repository.NewUserRepository(db.DB)

	// service
	authService := service.NewAuthService(authRepository, cfg, jwtAuth)
	categoryService := service.NewCategoryService(categoryRepository)
	contentService := service.NewContentService(contentRepository, cfg, r2Adapter)
	userService := service.NewUserService(userRepository)

	// handler
	authHandler := handler.NewAuthHandler(authService, cfg, jwtAuth)
	categoryHandler := handler.NewCategoryHandler(categoryService)
	contentHandler := handler.NewContentHandler(contentService)
	userHandler := handler.NewUserHandler(userService, cfg, jwtAuth)

	app := fiber.New(fiber.Config{
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
		BodyLimit:    6 * 1024 * 1024,
	})

	appEnv := strings.TrimSpace(cfg.App.AppEnv)
	if appEnv == "" {
		appEnv = strings.TrimSpace(os.Getenv("APP_ENV"))
	}

	allowOrigins := "https://blog.jonathangunawan.com"
	if strings.EqualFold(appEnv, "development") {
		allowOrigins = "http://localhost:3000"
	}

	app.Use(cors.New(cors.Config{
		AllowOrigins:     allowOrigins,
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
	}))
	app.Use(func(c *fiber.Ctx) error {
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("Referrer-Policy", "no-referrer")
		return c.Next()
	})
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${ip} ${status} - ${latency} ${method} ${path}\n",
	}))
	app.Use(limiter.New(limiter.Config{
		Max:        60,
		Expiration: 10 * time.Minute,
		LimitReached: func(c *fiber.Ctx) error {
			var errResp response.ErrorResponseDefault
			errResp.Meta.Status = false
			errResp.Meta.Message = "Too many requests"
			return c.Status(fiber.StatusTooManyRequests).JSON(errResp)
		},
	}))

	api := app.Group("/api")

	if strings.EqualFold(appEnv, "development") {
		filePath := "./docs/apispec.yaml"
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			log.Fatal().Msgf("Error swagger file not found: %v", err)
		}

		swaggerCfg := swagger.Config{
			BasePath: "/api",
			FilePath: filePath,
			Path:     "docs",
			Title:    "Swagger API Docs",
		}
		api.Use(swagger.New(swaggerCfg))
	}

	api.Post("/login", authHandler.Login)

	adminApp := api.Group("/admin")
	adminApp.Use(middlewareAuth.CheckOrigin(allowOrigins))
	adminApp.Use(middlewareAuth.CheckToken())
	adminApp.Post("/logout", authHandler.Logout)

	categoryApp := adminApp.Group("/categories")
	categoryApp.Get("/", categoryHandler.GetCategories)
	categoryApp.Post("/", categoryHandler.CreateCategory)
	categoryApp.Get("/:id", categoryHandler.GetCategoryByID)
	categoryApp.Put("/:id", categoryHandler.EditCategoryByID)
	categoryApp.Delete("/:id", categoryHandler.DeleteCategoryByID)

	contentApp := adminApp.Group("/contents")
	contentApp.Get("/", contentHandler.GetContents)
	contentApp.Post("/", contentHandler.CreateContent)
	contentApp.Get("/:id", contentHandler.GetContentByID)
	contentApp.Put("/:id", contentHandler.EditContentByID)
	contentApp.Delete("/:id", contentHandler.DeleteContentByID)
	contentApp.Post("/upload-image", contentHandler.UploadImageR2)

	userApp := adminApp.Group("/users")
	userApp.Get("/profile", userHandler.GetUserByID)
	userApp.Put("/update-password", userHandler.UpdatePassword)

	feApp := api.Group("/fe")
	feApp.Get("/categories", categoryHandler.GetCategoryFE)
	feApp.Get("/contents", contentHandler.GetContentWithQuery)
	feApp.Get("/contents/:id", contentHandler.GetContentDetail)

	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = strings.TrimSpace(cfg.App.AppPort)
	}
	if port == "" {
		port = "8080"
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := app.Listen(":" + port); err != nil {
			log.Error().Err(err).Msg("listen error")
		}
	}()

	<-quit
	log.Info().Msg("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Error().Err(err).Msg("forced shutdown")
	}

	log.Info().Msg("server exited cleanly")
}
