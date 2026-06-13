package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"

	"kenaito-notebook-server/handlers"
	"kenaito-notebook-server/services"
)

func main() {
	cfg := DefaultConfig()

	db, err := services.NewDatabase()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()
	log.Println("[Server] Database initialized")

	fileSvc, err := services.NewFileService(cfg.BaseURL)
	if err != nil {
		log.Fatalf("Failed to initialize file service: %v", err)
	}
	log.Println("[Server] File service initialized")

	exportSvc := services.NewExportService()
	authSvc := services.NewAuthService()
	emailSvc := services.NewEmailService(cfg.SMTP)
	verifySvc := services.NewVerificationService()
	log.Println("[Server] Services initialized")

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// CORS
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
			return
		}
		c.Next()
	})

	noteHandler := handlers.NewNoteHandler(db)
	fileHandler := handlers.NewFileHandler(fileSvc)
	exportHandler := handlers.NewExportHandler(db, exportSvc)
	authHandler := handlers.NewAuthHandler(db, authSvc, emailSvc, verifySvc)

	// 静态文件服务（图片、附件等）
	r.Static("/api/files/assets", fileSvc.GetAssetsPath())

	api := r.Group("/api")
	{
		// 认证路由（无需登录）
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/send-code", authHandler.SendCode)
			auth.POST("/verify-code", authHandler.VerifyCode)
			auth.POST("/reset-password", authHandler.ResetPassword)
		}

		// 需要登录的路由
		protected := api.Group("")
		protected.Use(handlers.AuthMiddleware(authSvc))
		{
			// 用户信息
			protected.GET("/auth/me", authHandler.GetMe)

			// 笔记管理
			notes := protected.Group("/notes")
			{
				notes.GET("/tree", noteHandler.GetNoteTree)
				notes.GET("/search", noteHandler.SearchNotes)
				notes.GET("/:id", noteHandler.GetNote)
				notes.POST("", noteHandler.CreateNote)
				notes.PUT("/:id", noteHandler.UpdateNote)
				notes.DELETE("/:id", noteHandler.DeleteNote)
			}

			// 回收站
			trash := protected.Group("/trash")
			{
				trash.GET("", noteHandler.GetTrashNotes)
				trash.POST("/:id/restore", noteHandler.RestoreNote)
				trash.DELETE("/:id", noteHandler.PermanentDeleteNote)
				trash.DELETE("", noteHandler.EmptyTrash)
			}

			// 文件管理
			files := protected.Group("/files")
			{
				files.POST("/image", fileHandler.SaveImage)
				files.POST("/upload", fileHandler.SaveFile)
				files.GET("/server-url", fileHandler.GetFileServerURL)
				files.GET("/assets-path", fileHandler.GetAssetsPath)
			}

			// 导出
			export := protected.Group("/export")
			{
				export.POST("/html", exportHandler.ExportHTML)
				export.POST("/markdown", exportHandler.ExportMarkdown)
				export.POST("/folder-zip/:folderId", exportHandler.ExportFolderZip)
				export.POST("/all-zip", exportHandler.ExportAllZip)
			}
		}
	}

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("[Server] Starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
