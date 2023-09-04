package checker

import (
	"context"
	"github.com/hashicorp/vault/sdk/helper/pointerutil"
	gc "github.com/keloran/go-config"
	"testing"

	"github.com/Nerzal/gocloak/v13"
	"github.com/keloran/go-config/keycloak"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	pb "github.com/todo-lists-app/protobufs/generated/id_checker/v1"
)

type MockGoCloak struct {
	mock.Mock
}

func (m *MockGoCloak) GetClient(ctx context.Context, cfg keycloak.Keycloak) *gocloak.GoCloak {
	args := m.Called(ctx, cfg)
	return args.Get(0).(*gocloak.GoCloak)
}

func (m *MockGoCloak) LoginClient(ctx context.Context, clientID, clientSecret, realm string) (*gocloak.JWT, error) {
	args := m.Called(ctx, clientID, clientSecret, realm)
	return args.Get(0).(*gocloak.JWT), args.Error(1)
}

func (m *MockGoCloak) GetUserByID(ctx context.Context, token, realm, userID string) (*gocloak.User, error) {
	args := m.Called(ctx, token, realm, userID)
	return args.Get(0).(*gocloak.User), args.Error(1)
}

func (m *MockGoCloak) RetrospectToken(ctx context.Context, token, clientID, clientSecret, realm string) (*gocloak.IntroSpectTokenResult, error) {
	args := m.Called(ctx, token, clientID, clientSecret, realm)
	return args.Get(0).(*gocloak.IntroSpectTokenResult), args.Error(1)
}

func TestCheckId(t *testing.T) {
	mockGoCloak := new(MockGoCloak)
	cfg := &gc.Config{
		Keycloak: keycloak.Keycloak{
			Client: "testClient",
			Secret: "testSecret",
			Realm:  "testRealm",
		},
	}

	mockGoCloak.On("GetClient", mock.Anything, mock.Anything).Return(&gocloak.GoCloak{})
	mockGoCloak.On("LoginClient", mock.Anything, "testClient", "testSecret", "testRealm").Return(&gocloak.JWT{AccessToken: "testToken"}, nil)
	mockGoCloak.On("GetUserByID", mock.Anything, "testToken", "testRealm", "testUserId").Return(&gocloak.User{}, nil)
	mockGoCloak.On("RetrospectToken", mock.Anything, "testAccessToken", "testClient", "testSecret", "testRealm").Return(&gocloak.IntroSpectTokenResult{Active: pointerutil.BoolPtr(true)}, nil)

	valid, err := CheckId(context.Background(), cfg, "testUserId", "testAccessToken", mockGoCloak)

	assert.Nil(t, err)
	assert.True(t, valid)
}

func TestServer_CheckId(t *testing.T) {
	mockGoCloak := new(MockGoCloak)
	server := &Server{
		Config:  &gc.Config{},
		GoCloak: mockGoCloak,
	}

	mockGoCloak.On("GetClient", mock.Anything, mock.Anything).Return(&gocloak.GoCloak{})
	mockGoCloak.On("LoginClient", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&gocloak.JWT{AccessToken: "testToken"}, nil)
	mockGoCloak.On("GetUserByID", mock.Anything, "testToken", mock.Anything, "testUserId").Return(&gocloak.User{}, nil)
	mockGoCloak.On("RetrospectToken", mock.Anything, "testAccessToken", mock.Anything, mock.Anything, mock.Anything).Return(&gocloak.IntroSpectTokenResult{Active: pointerutil.BoolPtr(true)}, nil)

	resp, err := server.CheckId(context.Background(), &pb.CheckIdRequest{
		Id:          "testUserId",
		AccessToken: "testAccessToken",
	})

	assert.Nil(t, err)
	assert.True(t, resp.IsValid)
}

// You can add more test cases, for example, when there's an error or when the user is not found.
