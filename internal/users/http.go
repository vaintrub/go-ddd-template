package main

import (
	"errors"
	"net"
	"net/http"

	"github.com/go-chi/render"
	"github.com/vaintrub/go-ddd-template/internal/common/auth"
	casdoorauth "github.com/vaintrub/go-ddd-template/internal/common/auth/casdoor"
	"github.com/vaintrub/go-ddd-template/internal/common/server/httperr"
)

type HttpServer struct {
	db      db
	casdoor *casdoorauth.Service
}

func (h HttpServer) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	authUser, err := auth.UserFromCtx(r.Context())
	if err != nil {
		httperr.RespondWithSlugError(err, w, r)
		return
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		err = h.db.UpdateLastIP(r.Context(), authUser.UUID, host)
		if err != nil {
			httperr.InternalError("internal-server-error", err, w, r)
			return
		}
	}

	user, err := h.db.GetUser(r.Context(), authUser.UUID)
	if err != nil {
		httperr.InternalError("cannot-get-user", err, w, r)
		return
	}

	userResponse := User{
		DisplayName: authUser.DisplayName,
		Balance:     user.Balance,
		Role:        authUser.Role,
	}

	render.Respond(w, r, userResponse)
}

func (h HttpServer) CasdoorCallback(w http.ResponseWriter, r *http.Request, params CasdoorCallbackParams) {
	if h.casdoor == nil {
		httperr.BadRequest("casdoor-not-configured", errors.New("casdoor integration is disabled"), w, r)
		return
	}

	callbackResult, err := h.casdoor.HandleCallback(params.Code, params.State)
	if err != nil {
		httperr.InternalError("casdoor-callback-failed", err, w, r)
		return
	}

	response := CasdoorOAuthResponse{
		AccessToken: callbackResult.AccessToken,
		TokenType:   callbackResult.TokenType,
		ExpiresIn:   callbackResult.ExpiresIn,
		User: CasdoorUser{
			Name: callbackResult.User.Name,
		},
	}

	if callbackResult.RefreshToken != nil {
		response.RefreshToken = callbackResult.RefreshToken
	}
	if callbackResult.User.Owner != nil {
		response.User.Owner = callbackResult.User.Owner
	}
	if callbackResult.User.DisplayName != nil {
		response.User.DisplayName = callbackResult.User.DisplayName
	}
	if callbackResult.User.Email != nil {
		response.User.Email = callbackResult.User.Email
	}
	if callbackResult.User.Avatar != nil {
		response.User.Avatar = callbackResult.User.Avatar
	}
	if callbackResult.User.Id != nil {
		response.User.Id = callbackResult.User.Id
	}

	render.Respond(w, r, response)
}
