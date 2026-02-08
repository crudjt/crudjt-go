package grpcserver

import (
    "context"
    tokenpb "github.com/VladAkymov/crudjt/proto"
)

type Server struct {
    tokenpb.UnimplementedTokenServiceServer
}

func (s *Server) CreateToken(ctx context.Context, req *tokenpb.CreateTokenRequest) (*tokenpb.CreateTokenResponse, error) {
    return &tokenpb.CreateTokenResponse{
        Token: "super-token-123",
    }, nil
}
