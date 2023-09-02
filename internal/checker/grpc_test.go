package checker

import (
	"context"
	"github.com/Nerzal/gocloak/v13"
	gc "github.com/keloran/go-config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	pb "github.com/todo-lists-app/protobufs/generated/id_checker/v1"
	"testing"
)

type MockGocloakClient struct {
	mock.Mock
}

func (m *MockGocloakClient) CheckId(ctx context.Context, config *gc.Config, id, accessToken string) (bool, error) {
	args := m.Called(ctx, config, id, accessToken)
	return args.Bool(0), args.Error(1)
}

func (m *MockGocloakClient) LoginClient(ctx context.Context, clientID, clientSecret, realm string) (*gocloak.JWT, error) {
	args := m.Called(ctx, clientID, clientSecret, realm)
	return args.Get(0).(*gocloak.JWT), args.Error(1)
}

func TestServer_CheckId(t *testing.T) {
	mockIdChecker := new(MockGocloakClient)
	mockIdChecker.On("CheckId", mock.Anything, mock.Anything, "testId", "testToken").Return(true, nil)

	server := &Server{
		Config: &gc.Config{},
	}

	req := &pb.CheckIdRequest{
		Id:          "testId",
		AccessToken: "testToken",
	}

	resp, err := server.CheckId(context.Background(), req)

	assert.NoError(t, err)
	assert.True(t, resp.IsValid)
}
