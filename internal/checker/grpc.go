package checker

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Nerzal/gocloak/v13"
	"github.com/bugfixes/go-bugfixes/logs"
	"github.com/go-resty/resty/v2"
	"github.com/keloran/go-config/keycloak"

	"github.com/hashicorp/vault/sdk/helper/pointerutil"
	gc "github.com/keloran/go-config"
	pb "github.com/todo-lists-app/protobufs/generated/id_checker/v1"
)

type Server struct {
	Config  *gc.Config
	GoCloak GocloakInterface

	pb.UnimplementedIdCheckerServiceServer
}

type GocloakInterface interface {
	GetClient(ctx context.Context, cfg keycloak.Keycloak) *gocloak.GoCloak
	LoginClient(ctx context.Context, clientID, clientSecret, realm string) (*gocloak.JWT, error)
	GetUserByID(ctx context.Context, token, realm, userID string) (*gocloak.User, error)
	RetrospectToken(ctx context.Context, token, clientID, clientSecret, realm string) (*gocloak.IntroSpectTokenResult, error)
}

type RealGoCloak struct {
	Client *gocloak.GoCloak
}

func (r *RealGoCloak) GetClient(ctx context.Context, cfg keycloak.Keycloak) *gocloak.GoCloak {
	client := gocloak.NewClient(cfg.Host)
	cond := func(resp *resty.Response, err error) bool {
		if resp != nil && resp.IsError() {
			if e, ok := resp.Error().(*gocloak.HTTPErrorResponse); ok {
				msg := e.String()
				logs.Infof("error: %s", msg)
				return strings.Contains(msg, "Cached clientScope not found") || strings.Contains(msg, "unknown_error")
			}
		}
		return false
	}
	rest := client.RestyClient()
	rest.SetRetryCount(10).SetRetryWaitTime(2 * time.Second).AddRetryCondition(cond)
	client.SetRestyClient(rest)
	return client
}
func (r *RealGoCloak) LoginClient(ctx context.Context, clientID, clientSecret, realm string) (*gocloak.JWT, error) {
	return r.Client.LoginClient(ctx, clientID, clientSecret, realm)
}
func (r *RealGoCloak) GetUserByID(ctx context.Context, token, realm, userID string) (*gocloak.User, error) {
	return r.Client.GetUserByID(ctx, token, realm, userID)
}
func (r *RealGoCloak) RetrospectToken(ctx context.Context, token, clientID, clientSecret, realm string) (*gocloak.IntroSpectTokenResult, error) {
	return r.Client.RetrospectToken(ctx, token, clientID, clientSecret, realm)
}

func (s *Server) CheckId(ctx context.Context, r *pb.CheckIdRequest) (*pb.CheckIdResponse, error) {
	validId, err := CheckId(ctx, s.Config, r.GetId(), r.GetAccessToken(), s.GoCloak)
	if err != nil {
		return &pb.CheckIdResponse{
			IsValid: false,
			Status:  pointerutil.StringPtr(fmt.Sprintf("failed to check id: %v", err)),
		}, err
	}

	if !validId {
		logs.Infof("id is not valid: %s", r.GetId())
		return &pb.CheckIdResponse{
			IsValid: false,
			Status:  pointerutil.StringPtr("id is not valid"),
		}, nil
	}

	return &pb.CheckIdResponse{
		IsValid: true,
	}, nil
}

func CheckId(ctx context.Context, cfg *gc.Config, userId, accessToken string, gc GocloakInterface) (bool, error) {
	client := gc.GetClient(ctx, cfg.Keycloak)
	if client == nil {
		return false, logs.Errorf("error getting client")
	}

	token, err := client.LoginClient(ctx, cfg.Keycloak.Client, cfg.Keycloak.Secret, cfg.Keycloak.Realm)
	if err != nil {
		return false, logs.Errorf("error logging in: %v", err)
	}
	user, err := client.GetUserByID(ctx, token.AccessToken, cfg.Keycloak.Realm, userId)
	if err != nil {
		return false, logs.Errorf("error getting user: %v", err)
	}
	if user == nil {
		logs.Infof("user not found: %s", userId)
		return false, nil
	}

	retroToken, err := client.RetrospectToken(ctx, accessToken, cfg.Keycloak.Client, cfg.Keycloak.Secret, cfg.Keycloak.Realm)
	if err != nil {
		return false, logs.Errorf("error introspecting token: %v", err)
	}

	if *retroToken.Active == false && token.ExpiresIn < 1 {
		return false, logs.Errorf("token is not active")
	}

	if *retroToken.Active == false && token.ExpiresIn > 1 {
		return true, nil
	}

	return false, logs.Error("something went wrong")
}
