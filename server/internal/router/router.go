package router

import (
	"crapp-go/internal/handlers"
	"crapp-go/internal/models"
	"crapp-go/views"
	"crapp-go/views/common"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Setup(log *zap.Logger, assessment *models.Assessment) *gin.Engine {
	// Set up a new Gin router, add recovery middleware and request logging.
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(RequestLogger(log))

	store := cookie.NewStore([]byte("secret"))
	router.Use(sessions.Sessions("mysession", store))
	// Use our new middleware on every request, passing the logger to it.
	router.Use(UserLoaderMiddleware(log))
	router.Static("/assets", "./assets")

	// Pass the logger and assessment model to the handlers
	authHandler := handlers.NewAuthHandler(log, assessment)
	assessmentHandler := handlers.NewAssessmentHandler(log, assessment)

	router.GET("/", func(c *gin.Context) {
		// We now check for the user object in the context, not the session.
		_, isLoggedIn := c.Get("user")
		isHTMX := c.GetHeader("HX-Request") == "true"

		if isLoggedIn {
			assessmentHandler.Start(c, isHTMX)
			return
		}

		// For guests, render the full page with the Login component.
		views.Page("CRAPP", false).Render(c.Request.Context(), c.Writer)
	})

	router.GET("/nav", func(c *gin.Context) {
		// Pass the result of the context check to the Nav component.
		_, isLoggedIn := c.Get("user")
		common.Nav(isLoggedIn).Render(c, c.Writer)
	})

	router.GET("/login", authHandler.ShowLoginPage)
	router.POST("/login", authHandler.Login)
	router.POST("/logout", authHandler.Logout)
	router.GET("/register", authHandler.RegisterPage)
	router.POST("/register", authHandler.Register)

	authorized := router.Group("/assessment")
	authorized.Use(AuthRequired(log))
	{
		// Pass false because this is always a full page load or redirect context
		authorized.GET("", func(c *gin.Context) {
			isHTMX := c.GetHeader("HX-Request") == "true"
			assessmentHandler.Start(c, isHTMX)
		})
		authorized.POST("/prev", assessmentHandler.PreviousQuestion)
		authorized.POST("/next", assessmentHandler.NextQuestion)
	}

	return router
}
