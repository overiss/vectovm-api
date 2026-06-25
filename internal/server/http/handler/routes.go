package hanlderHttp

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/overiss/vectovm-api/internal/server/http/middleware"

	_ "github.com/overiss/vectovm-api/api/docs"
)

func RegisterRoutes(router *gin.Engine, h *Container) {
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := router.Group("/api/v1")
	{
		v1.POST("/signup", h.Auth.SignUp)
		v1.POST("/auth/token", h.Auth.ExchangeToken)
		v1.POST("/auth/refresh", h.Auth.Refresh)
		v1.POST("/auth/logout", h.Auth.Logout)
	}

	authMiddleware := middleware.Auth(h.Verifier)
	protected := v1.Group("")
	protected.Use(authMiddleware)
	{
		protected.GET("/me", h.User.Me)

		datanodes := protected.Group("/datanodes")
		{
			datanodes.POST("", h.Datanode.Create)
			datanodes.GET("", h.Datanode.List)
			datanodes.POST("/vault/deploy", h.Datanode.DeployVault)
			datanodes.GET("/jobs/:id", h.Datanode.JobStatus)
			datanodes.GET("/:name/runtime", h.Datanode.Runtime)
		}

		vms := protected.Group("/vms")
		{
			vms.POST("", h.VM.Create)
			vms.GET("", h.VM.List)
			vms.GET("/:name", h.VM.Get)
		}
	}
}
