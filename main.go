package main

import (
	"marblegame/routes"
	"net/http"
	"os"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i any) error {
	if err := cv.validator.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

func main() {
	e := echo.New()

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Set headers to disable caching for static files in dev
			if os.Getenv("ENV") == "development" {
				c.Response().Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
				c.Response().Header().Set("Pragma", "no-cache")
				c.Response().Header().Set("Expires", "0")
				return next(c)
			}

			return next(c)
		}
	})

	e.Static("/", "static")

	e.Validator = &CustomValidator{validator: validator.New()}

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `${time_rfc3339} ${method} ${uri} ${status}
`,
	}))
	e.Use(middleware.Recover())

	routes.MarbleGameRouteHandler(e)
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		code := http.StatusInternalServerError
		if he, ok := err.(*echo.HTTPError); ok {
			code = he.Code
		}
		if code == http.StatusNotFound {
			c.String(http.StatusNotFound, "404 Not Found")
			return
		}
		e.DefaultHTTPErrorHandler(err, c)
	}

	e.Logger.Fatal(e.Start(":3000"))
}
