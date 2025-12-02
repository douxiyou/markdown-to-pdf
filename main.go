package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/douxiyou/md-to-pdf/chrome"
	"github.com/douxiyou/md-to-pdf/convert"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	port := flag.String("port", "3000", "port to listen on")
	flag.Parse()
	fmt.Println("port:", *port)

	// 初始化chrome
	chrome.InitGlobalChrome()
	defer chrome.GlobalCancel()

	// 启动健康检查协程
	go healthCheck()

	e := echo.New()

	// 使用Echo的内置日志中间件替代自定义中间件
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}, latency=${latency_human}\n",
	}))

	// 添加恢复中间件，防止程序因panic退出
	e.Use(middleware.Recover())

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	e.POST("/utils/convert-pdf", convert.HandleMarkdownToPdf)

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status":             "ok",
			"chrome_initialized": chrome.IsInitialized(),
			"last_used":          chrome.GetLastUsed(),
		})
	})

	e.RouteNotFound("/*", func(c echo.Context) error {
		requestUrl := c.Request().URL.String()
		return c.String(http.StatusNotFound, requestUrl+" Page not found")
	})

	// 设置服务器超时
	e.Server.ReadTimeout = 60 * time.Second
	e.Server.WriteTimeout = 60 * time.Second

	e.Logger.Fatal(e.Start(":" + *port))
}

// healthCheck 健康检查协程
func healthCheck() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if chrome.IsInitialized() {
			// 检查实例是否长时间未使用（超过1小时）
			if time.Since(chrome.GetLastUsed()) > time.Hour {
				fmt.Println("Chrome实例长时间未使用，重新初始化以保持活跃状态")
				chrome.Reset()
				chrome.InitGlobalChrome()
			}
		} else {
			fmt.Println("Chrome实例未初始化，重新初始化...")
			chrome.InitGlobalChrome()
		}
	}
}
