package checker

import (
	"context"
	"fmt"
	"github.com/hashicorp/vault/sdk/helper/pointerutil"
	"github.com/todo-lists-app/id-checker/internal/config"
	pb "github.com/todo-lists-app/protobufs/generated/id_checker/v1"
)

type Server struct {
	Config *config.Config
	pb.UnimplementedIdCheckerServiceServer
}

func (s *Server) CheckId(ctx context.Context, r *pb.CheckIdRequest) (*pb.CheckIdResponse, error) {
	validId, err := CheckId(ctx, s.Config, r.GetId(), r.GetAccessToken())
	if err != nil {
		return &pb.CheckIdResponse{
			IsValid: false,
			Status:  pointerutil.StringPtr(fmt.Sprintf("failed to check id: %v", err)),
		}, err
	}

	if !validId {
		return &pb.CheckIdResponse{
			IsValid: false,
			Status:  pointerutil.StringPtr("id is not valid"),
		}, nil
	}

	return &pb.CheckIdResponse{
		IsValid: true,
	}, nil
}
