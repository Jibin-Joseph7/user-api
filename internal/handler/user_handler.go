package handler

import (
	"errors"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/Jibin-Joseph7/user-api/internal/models"
	"github.com/Jibin-Joseph7/user-api/internal/repository"
	"github.com/Jibin-Joseph7/user-api/internal/service"
)

type UserHandler struct {
	svc      *service.UserService
	validate *validator.Validate
	log      *zap.Logger
}

func New(svc *service.UserService, log *zap.Logger) *UserHandler {
	return &UserHandler{svc: svc, validate: validator.New(), log: log}
}

func parseID(c *fiber.Ctx) (int32, bool) {
	id, err := strconv.ParseInt(c.Params("id"), 10, 32)
	if err != nil || id < 1 {
		_ = c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
		return 0, false
	}
	return int32(id), true
}

// bind parses and validates the body. It writes the error response itself and returns false on failure.
func (h *UserHandler) bind(c *fiber.Ctx, req *models.UserRequest) bool {
	if err := c.BodyParser(req); err != nil {
		_ = c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid JSON body"})
		return false
	}
	if err := h.validate.Struct(req); err != nil {
		details := map[string]string{}
		var verrs validator.ValidationErrors
		if errors.As(err, &verrs) {
			for _, fe := range verrs {
				details[fe.Field()] = "failed rule: " + fe.Tag()
			}
		}
		_ = c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "validation failed", "details": details})
		return false
	}
	return true
}

func (h *UserHandler) fail(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, service.ErrInvalidDOB), errors.Is(err, service.ErrFutureDOB):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	default:
		h.log.Error("internal error", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
}

func (h *UserHandler) Create(c *fiber.Ctx) error {
	var req models.UserRequest
	if !h.bind(c, &req) {
		return nil
	}
	resp, err := h.svc.Create(c.UserContext(), req)
	if err != nil {
		return h.fail(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(resp)
}

func (h *UserHandler) Get(c *fiber.Ctx) error {
	id, ok := parseID(c)
	if !ok {
		return nil
	}
	resp, err := h.svc.GetByID(c.UserContext(), id)
	if err != nil {
		return h.fail(c, err)
	}
	return c.JSON(resp)
}

func (h *UserHandler) Update(c *fiber.Ctx) error {
	id, ok := parseID(c)
	if !ok {
		return nil
	}
	var req models.UserRequest
	if !h.bind(c, &req) {
		return nil
	}
	resp, err := h.svc.Update(c.UserContext(), id, req)
	if err != nil {
		return h.fail(c, err)
	}
	return c.JSON(resp)
}

func (h *UserHandler) Delete(c *fiber.Ctx) error {
	id, ok := parseID(c)
	if !ok {
		return nil
	}
	if err := h.svc.Delete(c.UserContext(), id); err != nil {
		return h.fail(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// GET /users?page=1&limit=10  (total count is returned in the X-Total-Count header)
func (h *UserHandler) List(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	users, total, err := h.svc.List(c.UserContext(), page, limit)
	if err != nil {
		return h.fail(c, err)
	}
	c.Set("X-Total-Count", strconv.FormatInt(total, 10))
	return c.JSON(users)
}