package main

import (
	"log"

	"push_qq/core"
	"push_qq/router"

	"github.com/gofiber/fiber/v2"
)

func main() {
	log.SetFlags(log.Llongfile | log.LstdFlags) // 日志打印带上文件行

	// 初始化QQ机器人
	if err := core.InitQQBot(); err != nil {
		log.Fatal("初始化QQ机器人失败:", err)
	}

	app := fiber.New(fiber.Config{
		DisableStartupMessage:   true,  // 正式环境关闭启动信息
		EnableTrustedProxyCheck: false, // 开启代理检查
		BodyLimit:               6 * 1024 * 1024,
		ReadBufferSize:          8 * 1024, // 默认4096
	})

	router.SetupRoutes(app)

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	log.Fatal(app.Listen(":3206"))
}
