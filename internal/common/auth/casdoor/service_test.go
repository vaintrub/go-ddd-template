package casdoor

import (
	"errors"
	"testing"
	"time"

	"github.com/casdoor/casdoor-go-sdk/casdoorsdk"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

type fakeClient struct {
	token       *oauth2.Token
	claims      *casdoorsdk.Claims
	exchangeErr error
	parseErr    error
}

func (f *fakeClient) GetOAuthToken(code string, state string, opts ...casdoorsdk.OAuthOption) (*oauth2.Token, error) {
	if f.exchangeErr != nil {
		return nil, f.exchangeErr
	}
	return f.token, nil
}

func (f *fakeClient) ParseJwtToken(token string) (*casdoorsdk.Claims, error) {
	if f.parseErr != nil {
		return nil, f.parseErr
	}
	return f.claims, nil
}

func TestServiceHandleCallback(t *testing.T) {
	expires := time.Now().Add(90 * time.Minute)
	token := (&oauth2.Token{
		AccessToken:  "access",
		RefreshToken: "refresh",
		TokenType:    "bearer",
		Expiry:       expires,
	}).WithExtra(map[string]any{"expires_in": float64(5400)})

	claims := &casdoorsdk.Claims{
		User: casdoorsdk.User{
			Owner:       "org",
			Name:        "john",
			DisplayName: "John",
			Email:       "john@example.com",
			Avatar:      "http://avatar",
			Id:          "user-id",
		},
	}

	service := NewService(&fakeClient{
		token:  token,
		claims: claims,
	})

	resp, err := service.HandleCallback("code", "state")
	require.NoError(t, err)
	require.Equal(t, "access", resp.AccessToken)
	require.NotNil(t, resp.RefreshToken)
	require.Equal(t, "refresh", *resp.RefreshToken)
	require.Equal(t, "bearer", resp.TokenType)
	require.Equal(t, 5400, resp.ExpiresIn)
	require.Equal(t, "john", resp.User.Name)
	require.Equal(t, "John", *resp.User.DisplayName)
	require.Equal(t, "org", *resp.User.Owner)
	require.Equal(t, "john@example.com", *resp.User.Email)
	require.Equal(t, "http://avatar", *resp.User.Avatar)
	require.Equal(t, "user-id", *resp.User.Id)
}

func TestServiceHandleCallbackErrors(t *testing.T) {
	t.Run("not configured", func(t *testing.T) {
		var service *Service
		_, err := service.HandleCallback("code", "state")
		require.Error(t, err)
	})

	t.Run("exchange failure", func(t *testing.T) {
		service := NewService(&fakeClient{exchangeErr: errors.New("boom")})
		_, err := service.HandleCallback("code", "state")
		require.Error(t, err)
	})

	t.Run("parse failure", func(t *testing.T) {
		token := (&oauth2.Token{AccessToken: "token"}).WithExtra(map[string]any{"expires_in": 10})
		service := NewService(&fakeClient{
			token:    token,
			parseErr: errors.New("parse failed"),
		})

		_, err := service.HandleCallback("code", "state")
		require.Error(t, err)
	})
}

type fakeCertificateClient struct {
	app           *casdoorsdk.Application
	cert          *casdoorsdk.Cert
	appErr        error
	certErr       error
	requestedCert string
}

func (f *fakeCertificateClient) GetApplication(name string) (*casdoorsdk.Application, error) {
	if f.appErr != nil {
		return nil, f.appErr
	}
	return f.app, nil
}

func (f *fakeCertificateClient) GetCert(name string) (*casdoorsdk.Cert, error) {
	f.requestedCert = name
	if f.certErr != nil {
		return nil, f.certErr
	}
	return f.cert, nil
}

func TestResolveCertificate(t *testing.T) {
	t.Run("uses public key when available", func(t *testing.T) {
		client := &fakeCertificateClient{
			app: &casdoorsdk.Application{
				CertPublicKey: "PUBLIC",
			},
		}

		cert, err := resolveCertificate(client, "app")
		require.NoError(t, err)
		require.Equal(t, "PUBLIC", cert)
		require.Empty(t, client.requestedCert)
	})

	t.Run("fetches cert by name when needed", func(t *testing.T) {
		client := &fakeCertificateClient{
			app: &casdoorsdk.Application{
				Cert: "local-cert",
			},
			cert: &casdoorsdk.Cert{
				Certificate: "-----BEGIN CERTIFICATE-----\nabc\n-----END CERTIFICATE-----",
			},
		}

		cert, err := resolveCertificate(client, "app")
		require.NoError(t, err)
		require.Equal(t, client.cert.Certificate, cert)
		require.Equal(t, "local-cert", client.requestedCert)
	})
}
