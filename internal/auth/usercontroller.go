package auth

import (
	"context"
	"net/http"

	"github.com/impr0ver/uploaderGo/internal/servconfig"
	"github.com/impr0ver/uploaderGo/internal/serverstor"
	"github.com/labstack/echo/v4"
)

const (
	defaultCtxTimeout = servconfig.DefaultCtxTimeout
)

func RegisterUser(c echo.Context, cfg *servconfig.ServerConfig, memStor serverstor.MemoryStoragerInterface) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), defaultCtxTimeout)
	defer cancel()

	var user User

	if err := c.Bind(&user); err != nil {
		return c.Render(http.StatusOK, "register.html", map[string]interface{}{
			"title":   "Загрузка файлов",
			"message": "error: bind form",
		})
	}

	if user.Username == "" {
		return c.Render(http.StatusOK, "register.html", map[string]interface{}{
			"title":   "Загрузка файлов",
			"message": "error: username is empty",
		})
	}

	if user.Password == "" {
		return c.Render(http.StatusOK, "register.html", map[string]interface{}{
			"title":   "Загрузка файлов",
			"message": "error: password is empty",
		})
	}

	if user.Password != user.RePassword {
		return c.Render(http.StatusOK, "register.html", map[string]interface{}{
			"title":   "Загрузка файлов",
			"message": "error: passwords are not equal",
		})
	}

	if err := user.HashPassword(user.Password); err != nil {
		return c.Render(http.StatusOK, "register.html", map[string]interface{}{
			"title":   "Загрузка файлов",
			"message": "error - hash password error",
		})
	}

	err := memStor.RegisterNewUser(ctx, user.Username, user.Password)
	if err != nil {
		return c.Render(http.StatusOK, "register.html", map[string]interface{}{
			"title":   "Загрузка файлов",
			"message": "error - create user in db",
		})
	}

	c.Render(http.StatusOK, "register.html", map[string]interface{}{
		"title":    "Загрузка файлов",
		"username": user.Username,
	})
	return nil
}
