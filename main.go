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

	// 初始化chrome，最多重试5次
	var err error
	for i := 0; i < 5; i++ {
		err = chrome.InitGlobalChrome()
		if err == nil {
			break
		}
		fmt.Printf("第%d次初始化Chrome失败: %v\n", i+1, err)
		if i < 4 {
			// 等待一段时间再重试，逐渐增加等待时间
			waitTime := time.Duration(i+1) * 2 * time.Second
			fmt.Printf("等待 %v 后重试...\n", waitTime)
			time.Sleep(waitTime)
		}
	}

	if err != nil {
		fmt.Printf("警告: Chrome初始化失败，但服务将继续运行: %v\n", err)
		// 不再panic，而是继续运行服务，让后续请求触发初始化
		// panic(fmt.Sprintf("Failed to initialize Chrome after 5 attempts: %v", err))
	}

	// 注意：即使Chrome初始化失败，也不调用defer chrome.GlobalCancel()

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
		return c.String(http.StatusOK, "Hello, World! Service is running but Chrome may not be initialized yet.")
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
				err := chrome.InitGlobalChrome()
				if err != nil {
					fmt.Printf("Chrome实例重新初始化失败: %v\n", err)
				}
			}
		} else {
			fmt.Println("Chrome实例未初始化，重新初始化...")
			// 增加重试机制
			var err error
			for i := 0; i < 3; i++ {
				err = chrome.InitGlobalChrome()
				if err == nil {
					break
				}
				fmt.Printf("第%d次重新初始化Chrome失败: %v\n", i+1, err)
				if i < 2 {
					time.Sleep(2 * time.Second)
				}
			}
			if err != nil {
				fmt.Printf("Chrome实例初始化失败: %v\n", err)
			}
		}
	}
}
