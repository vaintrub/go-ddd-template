package casdoor

import (
	"fmt"
	"strconv"
	"time"

	"github.com/casdoor/casdoor-go-sdk/casdoorsdk"
	"github.com/vaintrub/go-ddd-template/internal/common/config"
	"golang.org/x/oauth2"
)

type casdoorClient interface {
	GetOAuthToken(code string, state string, opts ...casdoorsdk.OAuthOption) (*oauth2.Token, error)
	ParseJwtToken(token string) (*casdoorsdk.Claims, error)
}

type Service struct {
	client casdoorClient
}

func NewServiceFromConfig(cfg config.CasdoorConfig) (*Service, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	client := casdoorsdk.NewClient(
		cfg.Endpoint,
		cfg.ClientID,
		cfg.ClientSecret,
		cfg.Certificate,
		cfg.Organization,
		cfg.Application,
	)

	if client.Certificate == "" {
		cert, err := resolveCertificate(client, cfg.Application)
		if err != nil {
			return nil, err
		}
		client.Certificate = cert
	}

	return &Service{client: client}, nil
}

func NewService(client casdoorClient) *Service {
	if client == nil {
		return nil
	}
	return &Service{client: client}
}

func (s *Service) HandleCallback(code, state string) (*CallbackResult, error) {
	if s == nil || s.client == nil {
		return nil, fmt.Errorf("casdoor callback service is not configured")
	}

	token, err := s.client.GetOAuthToken(code, state)
	if err != nil {
		return nil, fmt.Errorf("exchange oauth token: %w", err)
	}

	claims, err := s.client.ParseJwtToken(token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("parse casdoor access token: %w", err)
	}

	return &CallbackResult{
		AccessToken:  token.AccessToken,
		RefreshToken: optionalString(token.RefreshToken),
		TokenType:    normalizeTokenType(token.TokenType),
		ExpiresIn:    extractExpiresInSeconds(token),
		User: CallbackUser{
			Owner:       optionalString(claims.Owner),
			Name:        claims.Name,
			DisplayName: optionalString(claims.DisplayName),
			Email:       optionalString(claims.Email),
			Avatar:      optionalString(claims.Avatar),
			Id:          optionalString(claims.Id),
		},
	}, nil
}

type certificateProvider interface {
	GetApplication(name string) (*casdoorsdk.Application, error)
	GetCert(name string) (*casdoorsdk.Cert, error)
}

func resolveCertificate(client certificateProvider, applicationName string) (string, error) {
	app, err := client.GetApplication(applicationName)
	if err != nil {
		return "", fmt.Errorf("fetch casdoor application: %w", err)
	}
	if app == nil {
		return "", fmt.Errorf("casdoor application %q not found", applicationName)
	}

	if app.CertPublicKey != "" {
		return app.CertPublicKey, nil
	}

	certName := app.Cert
	if certName == "" {
		return "", fmt.Errorf("casdoor application %q has no certificate assigned", applicationName)
	}

	cert, err := client.GetCert(certName)
	if err != nil {
		return "", fmt.Errorf("fetch casdoor certificate %q: %w", certName, err)
	}
	if cert == nil || cert.Certificate == "" {
		return "", fmt.Errorf("casdoor certificate %q is empty", certName)
	}

	return cert.Certificate, nil
}

type CallbackResult struct {
	AccessToken  string
	RefreshToken *string
	TokenType    string
	ExpiresIn    int
	User         CallbackUser
}

type CallbackUser struct {
	Owner       *string
	Name        string
	DisplayName *string
	Email       *string
	Avatar      *string
	Id          *string
}

func normalizeTokenType(tokenType string) string {
	if tokenType == "" {
		return "Bearer"
	}
	return tokenType
}

func extractExpiresInSeconds(token *oauth2.Token) int {
	if token == nil {
		return 0
	}

	if raw := token.Extra("expires_in"); raw != nil {
		switch v := raw.(type) {
		case int64:
			return int(v)
		case int32:
			return int(v)
		case float64:
			return int(v)
		case float32:
			return int(v)
		case string:
			if parsed, err := strconv.ParseFloat(v, 64); err == nil {
				return int(parsed)
			}
		}
	}

	if token.Expiry.IsZero() {
		return 0
	}

	remaining := time.Until(token.Expiry).Seconds()
	if remaining < 0 {
		return 0
	}

	return int(remaining)
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
