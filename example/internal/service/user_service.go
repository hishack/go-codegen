package service

import (
	"context"

	"codegen-example/internal/api"
	"codegen-example/internal/dal/vo"
)

type IUserService interface {
	CreateUser(ctx context.Context, req *api.CreateUserRequest) (vo.CreateUserVo, error)
	GetUser(ctx context.Context, userId int64) (vo.GetUserVo, error)
	ListUsers(ctx context.Context) (vo.ListUsersVo, error)
	UpdateUser(ctx context.Context, userId int64, req *api.UpdateUserRequest) (bool, error)
	PatchUser(ctx context.Context, userId int64, req *api.PatchUserRequest) (bool, error)
	DeleteUser(ctx context.Context, userId int64) (bool, error)
	GetUserProfileVo(ctx context.Context, id int64) (vo.GetUserProfileVo, error)
}
