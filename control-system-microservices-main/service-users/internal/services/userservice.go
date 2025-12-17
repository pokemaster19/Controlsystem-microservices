package services

import (
	"errors"
	"fmt"
	"regexp"

	"strings"

	"github.com/SpiritFoxo/control-system-microservices/service-users/internal/config"
	"github.com/SpiritFoxo/control-system-microservices/service-users/internal/models"
	"github.com/SpiritFoxo/control-system-microservices/service-users/internal/repositories"
	"github.com/SpiritFoxo/control-system-microservices/service-users/utils"
	"github.com/SpiritFoxo/control-system-microservices/shared/userroles"
)

type UserService struct {
	userRepo repositories.UserRepositoryInterface
	cfg      *config.Config
}

func NewUserService(userRepo repositories.UserRepositoryInterface, cfg *config.Config) *UserService {
	return &UserService{
		userRepo: userRepo,
		cfg:      cfg,
	}
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RegisterUserInput struct {
	Email    string   `json:"email"`
	Password string   `json:"password"`
	Name     string   `json:"name"`
	Roles    []string `json:"roles"`
}

type EditUserInput struct {
	Name  *string   `json:"name"`
	Roles *[]string `json:"roles"`
}

type UserListInput struct {
	Page        int    `json:"page"`
	Limit       int    `json:"limit"`
	EmailFilter string `json:"email_filter"`
	RoleFilter  string `json:"role_filter"`
}

type UserResponse struct {
	ID    uint     `json:"id"`
	Email string   `json:"email"`
	Name  string   `json:"name"`
	Roles []string `json:"roles"`
}

type UserListResult struct {
	Users      []UserResponse `json:"users"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	Limit      int            `json:"limit"`
	TotalPages int            `json:"total_pages"`
}

func isValidEmail(email string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}

func isValidRole(role string) bool {
	validRoles := []string{
		userroles.RoleEngineer,
		userroles.RoleManager,
		userroles.RoleObserver,
		userroles.RoleAdmin,
		userroles.RoleSuperadmin,
	}
	for _, validRole := range validRoles {
		if role == validRole {
			return true
		}
	}
	return false
}

func (s *UserService) RegisterUser(input RegisterUserInput) (*UserResponse, error) {

	if input.Email == "" || input.Password == "" || input.Name == "" {
		return nil, errors.New("email, password, and name are required")
	}
	if !isValidEmail(input.Email) {
		return nil, errors.New("invalid email format")
	}
	if len(input.Password) < 8 {
		return nil, errors.New("password must be at least 8 characters")
	}
	if len(input.Roles) == 0 {
		input.Roles = []string{userroles.RoleEngineer}
	}

	for _, role := range input.Roles {
		if !isValidRole(role) {
			return nil, fmt.Errorf("invalid role: %s", role)
		}
	}

	if _, err := s.userRepo.GetUserByEmail(input.Email); err == nil {
		return nil, errors.New("email already exists")
	}

	user := models.User{
		Email:    strings.ToLower(input.Email),
		Password: input.Password,
		Name:     input.Name,
		Roles:    input.Roles,
	}

	if err := user.HashPassword(); err != nil {
		return nil, fmt.Errorf("failed to hash password: %v", err)
	}

	if err := s.userRepo.CreateUser(&user); err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}

	return &UserResponse{
		ID:    user.ID,
		Email: user.Email,
		Name:  user.Name,
		Roles: user.Roles,
	}, nil
}

func (s *UserService) LoginUser(email, password string) (string, *UserResponse, error) {
	user, err := s.userRepo.GetUserByEmail(strings.ToLower(email))
	if err != nil {
		return "", nil, errors.New("invalid email or password")
	}

	if err := user.VerifyPassword(password); err != nil {
		return "", nil, errors.New("invalid email or password")
	}

	token, err := utils.GenerateToken(*user, s.cfg)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate JWT: %v", err)
	}

	return token, &UserResponse{
		ID:    user.ID,
		Email: user.Email,
		Name:  user.Name,
		Roles: user.Roles,
	}, nil
}

func (s *UserService) GetUserByID(id uint) (*UserResponse, error) {
	user, err := s.userRepo.GetUserByID(id)
	if err != nil {
		return nil, errors.New("user not found")
	}

	return &UserResponse{
		ID:    user.ID,
		Email: user.Email,
		Name:  user.Name,
		Roles: user.Roles,
	}, nil
}

func (s *UserService) UpdateUser(id uint, input EditUserInput) (*UserResponse, error) {
	user, err := s.userRepo.GetUserByID(id)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if input.Roles != nil {
		for _, role := range *input.Roles {
			if !isValidRole(role) {
				return nil, fmt.Errorf("invalid role: %s", role)
			}
		}
	}

	updates := make(map[string]interface{})
	if input.Name != nil && *input.Name != "" {
		updates["name"] = *input.Name
	}
	if input.Roles != nil {
		updates["roles"] = *input.Roles
	}

	if len(updates) == 0 {
		return &UserResponse{
			ID:    user.ID,
			Email: user.Email,
			Name:  user.Name,
			Roles: user.Roles,
		}, nil
	}

	if err := s.userRepo.UpdateUser(user, updates); err != nil {
		return nil, fmt.Errorf("failed to update user: %v", err)
	}

	return &UserResponse{
		ID:    user.ID,
		Email: user.Email,
		Name:  user.Name,
		Roles: user.Roles,
	}, nil
}

func (s *UserService) GetUsers(input UserListInput) (*UserListResult, error) {
	if input.Page < 1 {
		return nil, errors.New("invalid page number")
	}
	if input.Limit < 1 {
		return nil, errors.New("invalid limit value")
	}

	users, total, err := s.userRepo.GetUsers(input.Page, input.Limit, input.EmailFilter, input.RoleFilter)
	if err != nil {
		return nil, err
	}

	response := make([]UserResponse, 0, len(users))
	for _, user := range users {
		response = append(response, UserResponse{
			ID:    user.ID,
			Email: user.Email,
			Name:  user.Name,
			Roles: user.Roles,
		})
	}

	totalPages := int((total + int64(input.Limit) - 1) / int64(input.Limit))

	return &UserListResult{
		Users:      response,
		Total:      total,
		Page:       input.Page,
		Limit:      input.Limit,
		TotalPages: totalPages,
	}, nil
}
