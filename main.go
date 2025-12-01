package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/douxiyou/md-to-pdf/convert"
	"github.com/labstack/echo/v4"
)

func main() {
	port := flag.String("port", "3000", "port to listen on")
	flag.Parse()
	fmt.Println("port:", *port)
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})
	e.POST("/convert-pdf", convert.HandleMarkdownToPdf)
	e.Logger.Fatal(e.Start(":" + *port))
}
