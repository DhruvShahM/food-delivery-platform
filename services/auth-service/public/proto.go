package public

import (
	"food-delivery-platform/services/auth-service/internal/proto"
)

// Re-export proto types for integration testing
type (
	RegisterRequest  = proto.RegisterRequest
	RegisterResponse = proto.RegisterResponse
	LoginRequest     = proto.LoginRequest
	LoginResponse    = proto.LoginResponse
)

var (
	NewRegisterRequest  = func() *RegisterRequest { return &RegisterRequest{} }
	NewRegisterResponse = func() *RegisterResponse { return &RegisterResponse{} }
	NewLoginRequest     = func() *LoginRequest { return &LoginRequest{} }
	NewLoginResponse    = func() *LoginResponse { return &LoginResponse{} }
)
