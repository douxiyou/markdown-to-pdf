package main

import (
	"net/http"

	"github.com/douxiyou/md-to-pdf/convert"
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})
	e.POST("/convert-pdf", convert.HandleMarkdownToPdf)
	e.Logger.Fatal(e.Start(":3000"))
}
