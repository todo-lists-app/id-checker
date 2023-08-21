package checker

import (
	"context"
	"strings"
	"time"

	"github.com/Nerzal/gocloak/v13"
	"github.com/bugfixes/go-bugfixes/logs"
	"github.com/go-resty/resty/v2"
	gc "github.com/keloran/go-config"
)

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
