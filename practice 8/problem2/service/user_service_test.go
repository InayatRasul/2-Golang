package service

import (
	"errors"
	"testing"

	"go.uber.org/mock/gomock"
	"practice8/problem2/repository"
)

func TestRegisterUser(t *testing.T) {
	tests := []struct {
		name          string
		user          *repository.User
		email         string
		setupMock     func(*gomock.Controller) repository.UserRepository
		expectError   bool
		expectedError string
	}{
		{
			name:  "User already exists",
			user:  &repository.User{ID: 2, Name: "John"},
			email: "john@example.com",
			setupMock: func(ctrl *gomock.Controller) repository.UserRepository {
				mock := repository.NewMockUserRepository(ctrl)
				mock.EXPECT().GetByEmail("john@example.com").Return(&repository.User{ID: 1, Name: "John"}, nil)
				return mock
			},
			expectError:   true,
			expectedError: "user with this email already exists",
		},
		{
			name:  "New User Success",
			user:  &repository.User{ID: 2, Name: "Jane"},
			email: "jane@example.com",
			setupMock: func(ctrl *gomock.Controller) repository.UserRepository {
				mock := repository.NewMockUserRepository(ctrl)
				gomock.InOrder(
					mock.EXPECT().GetByEmail("jane@example.com").Return(nil, nil),
					mock.EXPECT().CreateUser(&repository.User{ID: 2, Name: "Jane"}).Return(nil),
				)
				return mock
			},
			expectError: false,
		},
		{
			name:  "Repository error on CreateUser",
			user:  &repository.User{ID: 3, Name: "Bob"},
			email: "bob@example.com",
			setupMock: func(ctrl *gomock.Controller) repository.UserRepository {
				mock := repository.NewMockUserRepository(ctrl)
				gomock.InOrder(
					mock.EXPECT().GetByEmail("bob@example.com").Return(nil, nil),
					mock.EXPECT().CreateUser(&repository.User{ID: 3, Name: "Bob"}).Return(errors.New("error on creation")),
				)
				return mock
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := tt.setupMock(ctrl)
			service := NewUserService(mockRepo)

			err := service.RegisterUser(tt.user, tt.email)

			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.expectError && err != nil && tt.expectedError != "" && err.Error() != tt.expectedError {
				t.Errorf("expected error %v but got %v", tt.expectedError, err.Error())
			}
		})
	}
}

func TestUpdateUserName(t *testing.T) {
	tests := []struct {
		name        string
		id          int
		newName     string
		setupMock   func(*gomock.Controller) repository.UserRepository
		expectError bool
	}{
		{
			name:    "Empty name",
			id:      1,
			newName: "",
			setupMock: func(ctrl *gomock.Controller) repository.UserRepository {
				mock := repository.NewMockUserRepository(ctrl)
				return mock
			},
			expectError: true,
		},
		{
			name:    "User not found",
			id:      999,
			newName: "NewName",
			setupMock: func(ctrl *gomock.Controller) repository.UserRepository {
				mock := repository.NewMockUserRepository(ctrl)
				mock.EXPECT().GetUserByID(999).Return(nil, errors.New("user not found"))
				return mock
			},
			expectError: true,
		},
		{
			name:    "Successful update",
			id:      1,
			newName: "UpdatedName",
			setupMock: func(ctrl *gomock.Controller) repository.UserRepository {
				mock := repository.NewMockUserRepository(ctrl)
				gomock.InOrder(
					mock.EXPECT().GetUserByID(1).Return(&repository.User{ID: 1, Name: "OldName"}, nil),
					mock.EXPECT().UpdateUser(&repository.User{ID: 1, Name: "UpdatedName"}).Return(nil),
				)
				return mock
			},
			expectError: false,
		},
		{
			name:    "UpdateUser fails",
			id:      2,
			newName: "NewName",
			setupMock: func(ctrl *gomock.Controller) repository.UserRepository {
				mock := repository.NewMockUserRepository(ctrl)
				gomock.InOrder(
					mock.EXPECT().GetUserByID(2).Return(&repository.User{ID: 2, Name: "OldName"}, nil),
					mock.EXPECT().UpdateUser(&repository.User{ID: 2, Name: "NewName"}).Return(errors.New("update failed")),
				)
				return mock
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := tt.setupMock(ctrl)
			service := NewUserService(mockRepo)

			err := service.UpdateUserName(tt.id, tt.newName)

			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestDeleteUser(t *testing.T) {
	tests := []struct {
		name        string
		id          int
		setupMock   func(*gomock.Controller) repository.UserRepository
		expectError bool
		expectedErr string
	}{
		{
			name: "Attempt to delete admin",
			id:   1,
			setupMock: func(ctrl *gomock.Controller) repository.UserRepository {
				mock := repository.NewMockUserRepository(ctrl)
				return mock
			},
			expectError: true,
			expectedErr: "it is not allowed to delete admin user",
		},
		{
			name: "Successful delete",
			id:   2,
			setupMock: func(ctrl *gomock.Controller) repository.UserRepository {
				mock := repository.NewMockUserRepository(ctrl)
				mock.EXPECT().DeleteUser(2).Return(nil)
				return mock
			},
			expectError: false,
		},
		{
			name: "Repository Error",
			id:   3,
			setupMock: func(ctrl *gomock.Controller) repository.UserRepository {
				mock := repository.NewMockUserRepository(ctrl)
				mock.EXPECT().DeleteUser(3).Return(errors.New("database error"))
				return mock
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := tt.setupMock(ctrl)
			service := NewUserService(mockRepo)

			err := service.DeleteUser(tt.id)

			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.expectError && err != nil && tt.expectedErr != "" && err.Error() != tt.expectedErr {
				t.Errorf("expected error message %q but got %q", tt.expectedErr, err.Error())
			}
		})
	}
}
