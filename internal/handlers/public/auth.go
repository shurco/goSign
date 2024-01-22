package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/shurco/gosign/internal/middleware"
	"github.com/shurco/gosign/internal/models"
	"github.com/shurco/gosign/pkg/logging"
	"github.com/shurco/gosign/pkg/security/password"
	"github.com/shurco/gosign/pkg/storage/redis"
	"github.com/shurco/gosign/pkg/utils/webutil"
)

// SignIn is ...
func SignIn(c *fiber.Ctx) error {
	log := logging.Log
	request := &models.SignIn{}

	if err := redis.Conn.Ping(); err != nil {
		log.Err(err).Send()
	}

	if err := c.BodyParser(request); err != nil {
		log.Err(err).Send()
		return webutil.StatusBadRequest(c, err.Error())
	}

	if err := request.Validate(); err != nil {
		log.Err(err).Send()
		return webutil.StatusBadRequest(c, err.Error())
	}

	// Temp data
	user := &models.User{
		ID:       uuid.New().String(),
		Name:     "Name",
		Email:    "name@mail.com",
		Password: "$2a$04$W/GFuzw3C6mZPB7RnNoPVutDu14WMKUiisI7FY852FvOoIFgSljfq",
	}

	if !password.ComparePasswords(user.Password, request.Password) {
		return webutil.StatusBadRequest(c, "Wrong Password")
	}

	token, err := middleware.CreateToken(user)
	if err != nil {
		return webutil.StatusInternalServerError(c)
	}

	return webutil.StatusOK(c, "Login Successfully", token)
}

// SignOut is ...
func SignOut(c *fiber.Ctx) error {
	return nil
}
