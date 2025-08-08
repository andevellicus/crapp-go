// server/internal/router/router.go
package router

import (
	"crapp-go/internal/config"
	"crapp-go/internal/handlers"
	"crapp-go/internal/models"
	"crapp-go/views"
	"crapp-go/views/common"
	"fmt"
	"net/http"
	"time"

	ratelimit "github.com/JGLTechnologies/gin-rate-limit"
	"github.com/a-h/templ"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/unrolled/secure"
	"go.uber.org/zap"
)

func keyFunc(c *gin.Context) string {
	return c.ClientIP()
}
func errorHandler(c *gin.Context, info ratelimit.Info) {
	c.String(429, "Too many requests. Try again later.")
}

func Setup(log *zap.Logger, assessment *models.Assessment) *gin.Engine {
	// Set up a new Gin router, add recovery middleware and request logging.
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(RequestLogger(log))

	store := cookie.NewStore([]byte(config.Conf.Server.SessionSecret))
	store.Options(sessions.Options{
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400 * 7,
	})
	router.Use(sessions.Sessions("mysession", store))

	// --- Now that sessions are initialized, other middleware can use them ---
	router.Use(NonceMiddleware())
	router.Use(CSRFProtection())
	router.Use(UserLoaderMiddleware(log))

	// The rest of the middleware
	router.Use(func(c *gin.Context) {
		isHTMX := c.GetHeader("HX-Request") == "true"
		if !isHTMX {
			nonce, _ := c.Get(CspNonceContextKey)
			csp := fmt.Sprintf(
				"script-src 'self' https://unpkg.com https://cdn.jsdelivr.net 'nonce-%s'; style-src 'self' https://fonts.googleapis.com 'unsafe-inline'; font-src 'self' https://fonts.gstatic.com",
				nonce,
			)
			c.Header("Content-Security-Policy", csp)
		}
		c.Next()
	})

	secureMiddleware := secure.New(secure.Options{
		FrameDeny:          true,
		ContentTypeNosniff: true,
		BrowserXssFilter:   true,
	})
	router.Use(func(c *gin.Context) {
		err := secureMiddleware.Process(c.Writer, c.Request)
		if err != nil {
			c.Abort()
			return
		}
	})

	router.Static("/assets", "./assets")

	// Handlers and routes
	authHandler := handlers.NewAuthHandler(log, assessment)
	assessmentHandler := handlers.NewAssessmentHandler(log, assessment)
	metricsHandler := handlers.NewMetricsHandler(log)
	resultsHandler := handlers.NewResultsHandler(log, assessment)
	userHandler := handlers.NewUserHandler(log)

	rateLimitStore := ratelimit.InMemoryStore(&ratelimit.InMemoryOptions{
		Rate:  time.Minute,
		Limit: 5,
	})
	limiter := ratelimit.RateLimiter(rateLimitStore, &ratelimit.Options{
		ErrorHandler: errorHandler,
		KeyFunc:      keyFunc,
	})

	router.GET("/", func(c *gin.Context) {
		_, isLoggedIn := c.Get("user")

		if isLoggedIn {
			assessmentHandler.Start(c, false)
			return
		}

		csrfToken, exists := c.Get("csrf_token")
		if !exists {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		cspNonce, _ := c.Get(CspNonceContextKey)

		loginComponent := views.Login(csrfToken.(string))
		views.Layout("CRAPP", false, csrfToken.(string), cspNonce.(string)).Render(templ.WithChildren(c.Request.Context(), loginComponent), c.Writer)
	})

	router.GET("/nav", func(c *gin.Context) {
		_, isLoggedIn := c.Get("user")
		common.Nav(isLoggedIn).Render(c, c.Writer)
	})

	router.GET("/login", authHandler.ShowLoginPage)
	router.POST("/login", limiter, authHandler.Login)
	router.POST("/logout", authHandler.Logout)
	router.GET("/register", authHandler.ShowRegisterPage)
	router.POST("/register", limiter, authHandler.Register)
	router.POST("/metrics", metricsHandler.SaveMetrics)

	authorized := router.Group("/")
	authorized.Use(AuthRequired(log))
	{
		assessmentRoutes := authorized.Group("/assessment")
		{
			assessmentRoutes.GET("", func(c *gin.Context) {
				isHTMX := c.GetHeader("HX-Request") == "true"
				assessmentHandler.Start(c, isHTMX)
			})
			assessmentRoutes.POST("/prev", assessmentHandler.PreviousQuestion)
			assessmentRoutes.POST("/next", assessmentHandler.NextQuestion)
			assessmentRoutes.GET("/results", resultsHandler.ShowResults)
		}

		profileRoutes := authorized.Group("/profile")
		{
			// Point both routes to the SAME handler
			profileRoutes.GET("", userHandler.ShowProfilePage)
			profileRoutes.GET("/:section", userHandler.ShowProfilePage)

			// POST routes for form submissions remain the same
			profileRoutes.POST("/update-info", userHandler.UpdateInfo)
			profileRoutes.POST("/update-password", userHandler.UpdatePassword)
			profileRoutes.POST("/delete", userHandler.DeleteAccount)
		}
	}

	return router
}
