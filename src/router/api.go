package router

import (
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/kevinanielsen/go-fast-cdn/src/database"
	"github.com/kevinanielsen/go-fast-cdn/src/handlers"
	authHandlers "github.com/kevinanielsen/go-fast-cdn/src/handlers/auth"
	dbHandlers "github.com/kevinanielsen/go-fast-cdn/src/handlers/db"
	fHandlers "github.com/kevinanielsen/go-fast-cdn/src/handlers/file"
	"github.com/kevinanielsen/go-fast-cdn/src/middleware"
	"github.com/kevinanielsen/go-fast-cdn/src/models"
	"github.com/kevinanielsen/go-fast-cdn/src/util"
)

func (s *Server) AddApiRoutes() {
	api := s.Engine.Group("/api")
	api.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, "pong")
	})

	// Authentication routes (public)
	authHandler := authHandlers.NewAuthHandler(database.NewUserRepo(database.DB))
	auth := api.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.RefreshToken)
		auth.POST("/logout", authHandler.Logout)
	}

	// Initialize auth middleware
	authMiddleware := middleware.NewAuthMiddleware()

	// Protected auth routes
	authProtected := api.Group("/auth")
	authProtected.Use(authMiddleware.RequireAuth())
	{
		authProtected.GET("/profile", authHandler.GetProfile)
		authProtected.PUT("/change-password", authHandler.ChangePassword)
		authProtected.PUT("/change-email", authHandler.ChangeEmail)
		authProtected.POST("/2fa", authHandler.Setup2FA)
		authProtected.POST("/2fa/verify", authHandler.Verify2FA)
	}

	cdn := api.Group("/cdn")
	fileHandler := fHandlers.NewFileHandler()

	// Public CDN routes (read-only)
	{
		cdn.GET("/size", handlers.GetSizeHandler)
		cdn.GET("/:type/all", fileHandler.HandleAll)
		cdn.GET("/:type/:filename", fileHandler.HandleMetadata)
		cdn.GET("/folder/:type", fileHandler.HandleFolderList)
		for name := range models.FileTypes {
			cdn.Group("/download/"+name, middleware.ServedFileHeaders()).
				Static("", filepath.Join(util.ExPath, "uploads", name))
		}
		cdn.GET("/dashboard", handlers.NewDashboardHandler(
			fileHandler.Repo("docs"),
			fileHandler.Repo("images"),
			fileHandler.Repo("audio"),
			database.NewUserRepo(database.DB),
			database.NewConfigRepo(database.DB),
		).GetDashboard)
	}

	// Protected CDN routes (require authentication)
	cdnProtected := cdn.Group("/")
	cdnProtected.Use(authMiddleware.RequireAuth())

	upload := cdnProtected.Group("upload")
	{
		upload.POST("/:type", fileHandler.HandleUpload)
	}

	delete := cdnProtected.Group("delete")
	{
		delete.DELETE("/:type/:filename", fileHandler.HandleDelete)
		// Bulk delete is a POST because it carries a body, which DELETE is not
		// reliably allowed to.
		delete.POST("/:type/bulk", fileHandler.HandleBulkDelete)
	}

	rename := cdnProtected.Group("rename")
	{
		rename.PUT("/:type", fileHandler.HandleRename)
	}

	folder := cdnProtected.Group("folder")
	{
		folder.POST("/:type", fileHandler.HandleFolderCreate)
		folder.DELETE("/:type", fileHandler.HandleFolderDelete)
	}

	resize := cdnProtected.Group("resize")
	{
		resize.PUT("/image", fHandlers.HandleImageResize)
	}
	// Admin-only routes
	adminRoutes := api.Group("/admin")
	adminRoutes.Use(authMiddleware.RequireAuth(), authMiddleware.RequireAdmin())
	{
		adminRoutes.POST("/drop/database", dbHandlers.HandleDropDB)

		adminUserHandler := authHandlers.NewAdminUserHandler(database.NewUserRepo(database.DB))
		{
			adminRoutes.GET("/users", adminUserHandler.ListUsers)
			adminRoutes.POST("/users", adminUserHandler.CreateUser)
			adminRoutes.PUT("/users/:id", adminUserHandler.UpdateUser)
			adminRoutes.DELETE("/users/:id", adminUserHandler.DeleteUser)
		}

		// Config endpoints (admin only)
		configHandler := handlers.NewConfigHandler(database.NewConfigRepo(database.DB))
		adminRoutes.GET("/config/registration", configHandler.GetRegistrationEnabled)
		adminRoutes.POST("/config/registration", configHandler.SetRegistrationEnabled)
		adminRoutes.GET("/config/access_token_ttl", configHandler.GetAccessTokenTTL)
		adminRoutes.POST("/config/access_token_ttl", configHandler.SetAccessTokenTTL)
	}

	// Public config endpoint for registration status
	configHandler := handlers.NewConfigHandler(database.NewConfigRepo(database.DB))
	api.GET("/config/registration", configHandler.GetRegistrationEnabled)
}
