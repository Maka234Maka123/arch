package middleware

import (
	"echo/internal/config"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/hertz-contrib/swagger"
	swaggerFiles "github.com/swaggo/files"

	// 导入生成的 Swagger 文档
	_ "echo/docs"
)

// RegisterSwagger 注册 Swagger 文档路由
// 需要在项目中运行 `swag init` 生成文档后，并导入 docs 包才能使用
func RegisterSwagger(h *server.Hertz, cfg *config.SwaggerConfig) {
	if !cfg.Enabled {
		return
	}

	// 注册 Swagger 路由
	// 访问 /swagger/index.html 查看文档
	h.GET(cfg.BasePath+"/*any", swagger.WrapHandler(
		swaggerFiles.Handler,
		swagger.URL(cfg.BasePath+"/doc.json"), // API 定义 JSON 路径
	))
}

// SwaggerHandler 返回 Swagger 处理器（如果需要作为中间件使用）
func SwaggerHandler(cfg *config.SwaggerConfig) app.HandlerFunc {
	return swagger.WrapHandler(
		swaggerFiles.Handler,
		swagger.URL(cfg.BasePath+"/doc.json"),
	)
}
