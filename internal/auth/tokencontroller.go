package auth

import (
	"context"
	"net/http"
	"time"

	"github.com/impr0ver/uploaderGo/internal/servconfig"
	"github.com/impr0ver/uploaderGo/internal/serverstor"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type TokenRequest struct {
	Username string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

type User struct {
	Username   string `form:"username" json:"username" binding:"required"`
	Password   string `form:"password" json:"password" binding:"required"`
	RePassword string `form:"repassword" json:"repassword" binding:"required"`
}

func (user *User) HashPassword(password string) error {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return err
	}
	user.Password = string(bytes)
	return nil
}

func (user *User) CheckPassword(providedPassword string) error {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(providedPassword))
	if err != nil {
		return err
	}
	return nil
}

func GenerateToken(c echo.Context, cfg *servconfig.ServerConfig, memStor serverstor.MemoryStoragerInterface) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), defaultCtxTimeout)
	defer cancel()

	var form TokenRequest
	var user User

	if err := c.Bind(&form); err != nil {
		return c.Render(http.StatusOK, "login.html", map[string]interface{}{
			"title": "Загрузка файлов",
			"msg":   "error: bad request",
		})
	}

	if form.Password == "" {
		return c.Render(http.StatusOK, "login.html", map[string]interface{}{
			"title":   "Загрузка файлов",
			"message": "error: password is empty",
		})
	}

	// check if users exists and password is correct
	dbUser, err := memStor.GetUserByName(ctx, form.Username)
	if err != nil {
		return c.Render(http.StatusOK, "login.html", map[string]interface{}{
			"title":   "Загрузка файлов",
			"message": "error: username not found",
		})
	}

	user = User{
		Username: dbUser.UserName,
		Password: dbUser.Password,
	}

	credentialError := user.CheckPassword(form.Password)
	if credentialError != nil {
		return c.Render(http.StatusOK, "login.html", map[string]interface{}{
			"title":   "Загрузка файлов",
			"message": "error: invalid credentials",
		})
	}

	tokenString, err := GenerateJWT(user.Username)
	if err != nil {
		return c.Render(http.StatusOK, "login.html", map[string]interface{}{
			"title":   "Загрузка файлов",
			"message": "error: generate JWT error",
		})
	}

	//create new auth cookie on 1 hour
	WriteCookie(c, "Authorization", tokenString, time.Now().Add(1*time.Hour), "/", c.Request().URL.Hostname(), false, false)
	c.Redirect(http.StatusFound, "/index") // "/secured/index"
	return nil
}

func WriteCookie(c echo.Context, name string, value string, expires time.Time, path string, domain string, secure bool, httpOnly bool) error {
	cookie := new(http.Cookie)
	cookie.Name = name
	cookie.Value = value
	cookie.Expires = expires
	cookie.Path = path
	cookie.Domain = domain
	cookie.Secure = secure
	cookie.HttpOnly = httpOnly
	c.SetCookie(cookie)
	return nil
}
