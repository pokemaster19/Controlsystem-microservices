package services

import (
	"testing"

	"github.com/SpiritFoxo/control-system-microservices/service-users/internal/config"
	"github.com/SpiritFoxo/control-system-microservices/service-users/internal/models"
	"github.com/SpiritFoxo/control-system-microservices/service-users/internal/repositories/mocks"
	"github.com/SpiritFoxo/control-system-microservices/service-users/utils"
	"github.com/SpiritFoxo/control-system-microservices/shared/userroles"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func newTestUser(id uint, email, name string, roles ...string) *models.User {
	return &models.User{
		Model: gorm.Model{ID: id},
		Email: email,
		Name:  name,
		Roles: roles,
	}
}

func setupTest(t *testing.T) (*UserService, *mocks.MockUserRepositoryInterface, func()) {
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockUserRepositoryInterface(ctrl)

	cfg := &config.Config{
		TokenSecret:         "test-secret",
		TokenMinuteLifespan: "5",
	}

	service := NewUserService(mockRepo, cfg)
	return service, mockRepo, ctrl.Finish
}

func TestUserService_RegisterUser(t *testing.T) {
	service, mockRepo, finish := setupTest(t)
	defer finish()

	hashed, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	tests := []struct {
		name        string
		input       RegisterUserInput
		setupMock   func()
		expected    *UserResponse
		expectedErr string
	}{
		{
			name: "успешная регистрация",
			input: RegisterUserInput{
				Email:    "test@example.com",
				Password: "password123",
				Name:     "Test User",
				Roles:    []string{userroles.RoleEngineer},
			},
			setupMock: func() {
				mockRepo.EXPECT().
					GetUserByEmail("test@example.com").
					Return((*models.User)(nil), assert.AnError)

				mockRepo.EXPECT().
					CreateUser(gomock.Any()).
					DoAndReturn(func(u *models.User) error {
						u.ID = 1
						u.Password = string(hashed)
						return nil
					})
			},
			expected: &UserResponse{
				ID:    1,
				Email: "test@example.com",
				Name:  "Test User",
				Roles: []string{userroles.RoleEngineer},
			},
		},
		{
			name: "пустое имя",
			input: RegisterUserInput{
				Email:    "test@example.com",
				Password: "password123",
				Name:     "",
			},
			setupMock:   func() {},
			expected:    nil,
			expectedErr: "email, password, and name are required",
		},
		{
			name: "невалидный email",
			input: RegisterUserInput{
				Email:    "invalid",
				Password: "password123",
				Name:     "Test",
			},
			setupMock:   func() {},
			expected:    nil,
			expectedErr: "invalid email format",
		},
		{
			name: "короткий пароль",
			input: RegisterUserInput{
				Email:    "test@example.com",
				Password: "123",
				Name:     "Test",
			},
			setupMock:   func() {},
			expected:    nil,
			expectedErr: "password must be at least 8 characters",
		},
		{
			name: "email уже существует",
			input: RegisterUserInput{
				Email:    "exists@example.com",
				Password: "password123",
				Name:     "Test",
			},
			setupMock: func() {
				mockRepo.EXPECT().
					GetUserByEmail("exists@example.com").
					Return(&models.User{Email: "exists@example.com"}, nil)
			},
			expected:    nil,
			expectedErr: "email already exists",
		},
		{
			name: "невалидная роль",
			input: RegisterUserInput{
				Email:    "test@example.com",
				Password: "password123",
				Name:     "Test",
				Roles:    []string{"invalid_role"},
			},
			setupMock:   func() {},
			expected:    nil,
			expectedErr: "invalid role: invalid_role",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			got, err := service.RegisterUser(tt.input)

			if tt.expectedErr != "" {
				assert.ErrorContains(t, err, tt.expectedErr)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}

func TestUserService_LoginUser(t *testing.T) {
	service, mockRepo, finish := setupTest(t)
	defer finish()

	hashed, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	user := newTestUser(1, "test@example.com", "Test User", userroles.RoleEngineer)
	user.Password = string(hashed)

	tests := []struct {
		name        string
		email       string
		password    string
		setupMock   func()
		wantToken   bool
		expected    *UserResponse
		expectedErr string
	}{
		{
			name:     "успешный вход",
			email:    "test@example.com",
			password: "password123",
			setupMock: func() {
				mockRepo.EXPECT().
					GetUserByEmail("test@example.com").
					Return(user, nil)
			},
			wantToken: true,
			expected: &UserResponse{
				ID:    1,
				Email: "test@example.com",
				Name:  "Test User",
				Roles: []string{userroles.RoleEngineer},
			},
		},
		{
			name:     "неверный пароль",
			email:    "test@example.com",
			password: "wrong",
			setupMock: func() {
				mockRepo.EXPECT().
					GetUserByEmail("test@example.com").
					Return(user, nil)
			},
			expectedErr: "invalid email or password",
		},
		{
			name:     "пользователь не найден",
			email:    "unknown@example.com",
			password: "password123",
			setupMock: func() {
				mockRepo.EXPECT().
					GetUserByEmail("unknown@example.com").
					Return((*models.User)(nil), assert.AnError)
			},
			expectedErr: "invalid email or password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			token, got, err := service.LoginUser(tt.email, tt.password)

			if tt.expectedErr != "" {
				assert.ErrorContains(t, err, tt.expectedErr)
				assert.Empty(t, token)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, got)

				if tt.wantToken {
					assert.NotEmpty(t, token)

					parsed, err := utils.ParseToken(token, service.cfg)
					assert.NoError(t, err)
					assert.True(t, parsed.Valid)

					claims, ok := parsed.Claims.(jwt.MapClaims)
					assert.True(t, ok)
					assert.Equal(t, float64(1), claims["id"])
					assert.Contains(t, claims["roles"], userroles.RoleEngineer)
				}
			}
		})
	}
}

func TestUserService_GetUserByID(t *testing.T) {
	service, mockRepo, finish := setupTest(t)
	defer finish()

	user := newTestUser(1, "test@example.com", "Test User", userroles.RoleEngineer)

	tests := []struct {
		name        string
		id          uint
		setupMock   func()
		expected    *UserResponse
		expectedErr string
	}{
		{
			name: "успешно",
			id:   1,
			setupMock: func() {
				mockRepo.EXPECT().GetUserByID(uint(1)).Return(user, nil)
			},
			expected: &UserResponse{
				ID:    1,
				Email: "test@example.com",
				Name:  "Test User",
				Roles: []string{userroles.RoleEngineer},
			},
		},
		{
			name: "не найден",
			id:   999,
			setupMock: func() {
				mockRepo.EXPECT().GetUserByID(uint(999)).Return((*models.User)(nil), assert.AnError)
			},
			expectedErr: "user not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			got, err := service.GetUserByID(tt.id)

			if tt.expectedErr != "" {
				assert.ErrorContains(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}

func TestUserService_UpdateUser(t *testing.T) {
	service, mockRepo, finish := setupTest(t)
	defer finish()

	user := newTestUser(1, "test@example.com", "Oldname", userroles.RoleEngineer)
	UpdatedName := "Newname"

	tests := []struct {
		name        string
		input       EditUserInput
		setupMock   func()
		expected    *UserResponse
		expectedErr string
	}{
		{
			name: "обновление имени",
			input: EditUserInput{
				Name: &UpdatedName,
			},
			setupMock: func() {
				mockRepo.EXPECT().GetUserByID(uint(1)).Return(user, nil)
				mockRepo.EXPECT().UpdateUser(user, gomock.Any()).DoAndReturn(func(u *models.User, updates map[string]interface{}) error {
					if name, ok := updates["name"]; ok {
						u.Name = name.(string)
					}
					if roles, ok := updates["roles"]; ok {
						u.Roles = roles.([]string)
					}
					return nil
				})
			},
			expected: &UserResponse{
				ID:    1,
				Email: "test@example.com",
				Name:  "Newname",
				Roles: []string{userroles.RoleEngineer},
			},
		},
		{
			name: "невалидная роль",
			input: EditUserInput{
				Roles: &[]string{"invalid_role"},
			},
			setupMock: func() {
				mockRepo.EXPECT().GetUserByID(uint(1)).Return(user, nil)
			},
			expectedErr: "invalid role: invalid_role",
		},
		{
			name: "пустое обновление",

			input: EditUserInput{},
			setupMock: func() {
				mockRepo.EXPECT().GetUserByID(uint(1)).Return(user, nil)
			},
			expected: &UserResponse{
				ID:    1,
				Email: "test@example.com",
				Name:  "Newname",
				Roles: []string{userroles.RoleEngineer},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			got, err := service.UpdateUser(1, tt.input)

			if tt.expectedErr != "" {
				assert.ErrorContains(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.Name, got.Name)
				if tt.input.Roles != nil {
					assert.Equal(t, tt.input.Roles, got.Roles)
				}
			}
		})
	}
}

func TestUserService_GetUsers(t *testing.T) {
	service, mockRepo, finish := setupTest(t)
	defer finish()

	users := []models.User{
		*newTestUser(1, "a@example.com", "A", userroles.RoleEngineer),
		*newTestUser(2, "b@example.com", "B", userroles.RoleManager),
	}

	tests := []struct {
		name        string
		input       UserListInput
		setupMock   func()
		expectedLen int
		expectedErr string
	}{
		{
			name:  "успешно",
			input: UserListInput{Page: 1, Limit: 10},
			setupMock: func() {
				mockRepo.EXPECT().GetUsers(1, 10, "", "").Return(users, int64(2), nil)
			},
			expectedLen: 2,
		},
		{
			name:        "невалидная страница",
			input:       UserListInput{Page: 0, Limit: 10},
			setupMock:   func() {},
			expectedErr: "invalid page number",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			result, err := service.GetUsers(tt.input)

			if tt.expectedErr != "" {
				assert.ErrorContains(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result.Users, tt.expectedLen)
				assert.Equal(t, int64(2), result.Total)
			}
		})
	}
}
