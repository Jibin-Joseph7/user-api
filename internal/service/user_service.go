package service

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"

	sqlc "github.com/Jibin-Joseph7/user-api/db/sqlc"
	"github.com/Jibin-Joseph7/user-api/internal/models"
	"github.com/Jibin-Joseph7/user-api/internal/repository"
)

const dateLayout = "2006-01-02"

var ErrInvalidDOB = errors.New("dob must be a valid date in YYYY-MM-DD format")
var ErrFutureDOB = errors.New("dob cannot be in the future")

type UserService struct {
	repo *repository.UserRepository
	log  *zap.Logger
}

func New(repo *repository.UserRepository, log *zap.Logger) *UserService {
	return &UserService{repo: repo, log: log}
}

// CalculateAge returns the age in full years as of `now`.
func CalculateAge(dob, now time.Time) int {
	age := now.Year() - dob.Year()
	if now.Month() < dob.Month() || (now.Month() == dob.Month() && now.Day() < dob.Day()) {
		age--
	}
	return age
}

func parseDOB(s string) (time.Time, error) {
	dob, err := time.Parse(dateLayout, s)
	if err != nil {
		return time.Time{}, ErrInvalidDOB
	}
	if dob.After(time.Now()) {
		return time.Time{}, ErrFutureDOB
	}
	return dob, nil
}

func toResponse(u sqlc.User, withAge bool) models.UserResponse {
	dob := u.Dob
	resp := models.UserResponse{ID: u.ID, Name: u.Name, Dob: dob.Format(dateLayout)}
	if withAge {
		age := CalculateAge(dob, time.Now())
		resp.Age = &age
	}
	return resp
}

func (s *UserService) Create(ctx context.Context, req models.UserRequest) (models.UserResponse, error) {
	dob, err := parseDOB(req.Dob)
	if err != nil {
		return models.UserResponse{}, err
	}
	u, err := s.repo.Create(ctx, req.Name, dob)
	if err != nil {
		return models.UserResponse{}, err
	}
	s.log.Info("user created", zap.Int32("id", u.ID))
	return toResponse(u, false), nil
}

func (s *UserService) GetByID(ctx context.Context, id int32) (models.UserResponse, error) {
	u, err := s.repo.Get(ctx, id)
	if err != nil {
		return models.UserResponse{}, err
	}
	s.log.Info("user fetched", zap.Int32("id", id))
	return toResponse(u, true), nil
}

func (s *UserService) Update(ctx context.Context, id int32, req models.UserRequest) (models.UserResponse, error) {
	dob, err := parseDOB(req.Dob)
	if err != nil {
		return models.UserResponse{}, err
	}
	u, err := s.repo.Update(ctx, id, req.Name, dob)
	if err != nil {
		return models.UserResponse{}, err
	}
	s.log.Info("user updated", zap.Int32("id", id))
	return toResponse(u, false), nil
}

func (s *UserService) Delete(ctx context.Context, id int32) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	s.log.Info("user deleted", zap.Int32("id", id))
	return nil
}

func (s *UserService) List(ctx context.Context, page, limit int) ([]models.UserResponse, int64, error) {
	users, total, err := s.repo.List(ctx, int32(limit), int32((page-1)*limit))
	if err != nil {
		return nil, 0, err
	}
	out := make([]models.UserResponse, 0, len(users))
	for _, u := range users {
		out = append(out, toResponse(u, true))
	}
	s.log.Info("users listed", zap.Int("count", len(out)), zap.Int("page", page))
	return out, total, nil
}