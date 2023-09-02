package checker

import (
	"context"
	"fmt"
	"github.com/Nerzal/gocloak/v13"
	"github.com/bugfixes/go-bugfixes/logs"
	"github.com/go-resty/resty/v2"
	"strings"
	"time"

	"github.com/hashicorp/vault/sdk/helper/pointerutil"
	gc "github.com/keloran/go-config"
	pb "github.com/todo-lists-app/protobufs/generated/id_checker/v1"
)

type Server struct {
	Config *gc.Config
	pb.UnimplementedIdCheckerServiceServer
}

type GocloakInterface interface {
	LoginClient(ctx context.Context, clientID, clientSecret, realm string) (*gocloak.JWT, error)
	GetUserByID(ctx context.Context, token, realm, userID string) (*gocloak.User, error)
	RetrospectToken(ctx context.Context, token, clientID, clientSecret, realm string) (*gocloak.IntroSpectTokenResult, error)
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

func CheckId(ctx context.Context, cfg *gc.Config, userId, accessToken string) (bool, error) {
	client := gocloak.NewClient(cfg.Keycloak.Host)
	cond := func(resp *resty.Response, err error) bool {
		if resp != nil && resp.IsError() {
			if e, ok := resp.Error().(*gocloak.HTTPErrorResponse); ok {
				msg := e.String()
				return strings.Contains(msg, "Cached clientScope not found") || strings.Contains(msg, "unknown_error")
			}
		}
		return false
	}
	rest := client.RestyClient()
	rest.SetRetryCount(10).SetRetryWaitTime(2 * time.Second).AddRetryCondition(cond)
	token, err := client.LoginClient(ctx, cfg.Keycloak.Client, cfg.Keycloak.Secret, cfg.Keycloak.Realm)
	if err != nil {
		return false, logs.Errorf("error logging in: %v", err)
	}
	user, err := client.GetUserByID(ctx, token.AccessToken, cfg.Keycloak.Realm, userId)
	if err != nil {
		return false, logs.Errorf("error getting user: %v", err)
	}
	if user == nil {
		return false, nil
	}

	retroToken, err := client.RetrospectToken(ctx, accessToken, cfg.Keycloak.Client, cfg.Keycloak.Secret, cfg.Keycloak.Realm)
	if err != nil {
		return false, logs.Errorf("error introspecting token: %v", err)
	}

	return *retroToken.Active, nil
}
