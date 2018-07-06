package users

import "context"

// DefaultUsersServer implements the Users service
type DefaultUsersServer struct{}

// Get implements the User service Get method
func (s *DefaultUsersServer) Get(ctx context.Context, req *GetUserReq) (*User, error) {
	return nil, nil
}

// Create implements the User service Create method
func (s *DefaultUsersServer) Create(ctx context.Context, req *CreateUserReq) (*User, error) {
	return nil, nil
}

// Login implements the User service Login method
func (s *DefaultUsersServer) Login(ctx context.Context, req *LoginUserReq) (*LoginUserResp, error) {
	return nil, nil
}
