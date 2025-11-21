package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/casdoor/casdoor-go-sdk/casdoorsdk"
)

type cfg struct {
	Endpoint     string
	Org          string
	App          string
	ClientID     string
	ClientSecret string
	CallbackURL  string
	CertName     string

	BootstrapClientID string
	BootstrapSecret   string
	BootstrapApp      string
}

func main() {
	log.SetFlags(0)

	c := load()

	httpClient := &http.Client{Timeout: 10 * time.Second}
	casdoorsdk.SetHttpClient(httpClient)
	waitForHealth(httpClient, c.Endpoint, 2*time.Minute)

	client := casdoorsdk.NewClient(c.Endpoint, c.BootstrapClientID, c.BootstrapSecret, "", c.Org, c.BootstrapApp)

	ensureOrganization(client, c)
	cert := ensureCertificate(client, c)
	ensureApplication(client, c, cert.Name)

	log.Println("casdoor init: done")
}

func load() cfg {
	c := cfg{
		Endpoint:          strings.TrimRight(os.Getenv("CASDOOR_ENDPOINT"), "/"),
		Org:               env("CASDOOR_ORG_SLUG", "CASDOOR_ORGANIZATION"),
		App:               os.Getenv("CASDOOR_APPLICATION_NAME"),
		ClientID:          os.Getenv("CASDOOR_CLIENT_ID"),
		ClientSecret:      os.Getenv("CASDOOR_CLIENT_SECRET"),
		CallbackURL:       os.Getenv("CASDOOR_CALLBACK_URL"),
		CertName:          os.Getenv("CASDOOR_CERTIFICATE_NAME"),
		BootstrapClientID: os.Getenv("CASDOOR_BOOTSTRAP_CLIENT_ID"),
		BootstrapSecret:   os.Getenv("CASDOOR_BOOTSTRAP_CLIENT_SECRET"),
		BootstrapApp:      os.Getenv("CASDOOR_BOOTSTRAP_APPLICATION"),
	}

	if c.CertName == "" && c.App != "" {
		c.CertName = c.App + "-cert"
	}

	if c.BootstrapClientID == "" {
		c.BootstrapClientID = c.ClientID
	}
	if c.BootstrapSecret == "" {
		c.BootstrapSecret = c.ClientSecret
	}
	if c.BootstrapApp == "" {
		c.BootstrapApp = c.App
	}

	return c
}

func env(keys ...string) string {
	for _, k := range keys {
		if v := strings.TrimSpace(os.Getenv(k)); v != "" {
			return v
		}
	}
	return ""
}

func waitForHealth(client *http.Client, endpoint string, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	url := strings.TrimRight(endpoint, "/") + "/api/health"

	for {
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
		resp, err := client.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			_ = resp.Body.Close()
			return
		}
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
		select {
		case <-ctx.Done():
			log.Fatalf("casdoor is not ready: %v", ctx.Err())
		case <-time.After(2 * time.Second):
		}
	}
}

func ensureOrganization(client *casdoorsdk.Client, c cfg) {
	org, err := client.GetOrganization(c.Org)
	if err == nil && org != nil {
		return
	}

	newOrg := &casdoorsdk.Organization{
		Owner:              "admin",
		Name:               c.Org,
		CreatedTime:        casdoorsdk.GetCurrentTime(),
		DisplayName:        c.Org,
		PasswordType:       "plain",
		DefaultApplication: c.App,
	}

	if _, err := client.AddOrganization(newOrg); err != nil {
		log.Fatalf("create organization: %v", err)
	}
}

func ensureCertificate(client *casdoorsdk.Client, c cfg) *casdoorsdk.Cert {
	cert, _ := client.GetCert(c.CertName)
	if cert != nil && cert.Certificate != "" && cert.PrivateKey != "" {
		return cert
	}

	pub, priv := makeCert(c)
	now := casdoorsdk.GetCurrentTime()

	cert = &casdoorsdk.Cert{
		Owner:           c.Org,
		Name:            c.CertName,
		CreatedTime:     now,
		DisplayName:     c.CertName,
		Scope:           "JWT",
		Type:            "x509",
		CryptoAlgorithm: "RS256",
		BitSize:         4096,
		ExpireInYears:   10,
		Certificate:     pub,
		PrivateKey:      priv,
	}

	if _, err := client.AddCert(cert); err != nil {
		log.Fatalf("create cert: %v", err)
	}
	return cert
}

func ensureApplication(client *casdoorsdk.Client, c cfg, certName string) {
	app, _ := client.GetApplication(c.App)
	if app == nil {
		app = &casdoorsdk.Application{
			Owner:              "admin",
			Name:               c.App,
			CreatedTime:        casdoorsdk.GetCurrentTime(),
			DisplayName:        c.App,
			Organization:       c.Org,
			Cert:               certName,
			GrantTypes:         []string{"authorization_code", "refresh_token"},
			RedirectUris:       []string{c.CallbackURL},
			ClientId:           c.ClientID,
			ClientSecret:       c.ClientSecret,
			TokenFormat:        "JWT",
			TokenSigningMethod: "RS256",
			EnablePassword:     true,
			EnableSignUp:       true,
		}
		if _, err := client.AddApplication(app); err != nil {
			log.Fatalf("create app: %v", err)
		}
		return
	}

	// Minimal sync of critical fields.
	app.Organization = c.Org
	app.Cert = certName
	app.ClientId = c.ClientID
	app.ClientSecret = c.ClientSecret
	app.RedirectUris = []string{c.CallbackURL}
	app.TokenFormat = "JWT"
	app.TokenSigningMethod = "RS256"
	if _, err := client.UpdateApplication(app); err != nil {
		log.Fatalf("update app: %v", err)
	}
}

func makeCert(c cfg) (string, string) {
	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		log.Fatalf("generate key: %v", err)
	}

	now := time.Now()
	template := &x509.Certificate{
		SerialNumber: big.NewInt(now.UnixNano()),
		Subject: pkix.Name{
			CommonName:   fmt.Sprintf("%s.%s", c.App, c.Org),
			Organization: []string{c.Org},
		},
		NotBefore:             now.Add(-time.Hour),
		NotAfter:              now.Add(10 * 365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		log.Fatalf("create cert: %v", err)
	}

	pub := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	priv := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	return string(pub), string(priv)
}
