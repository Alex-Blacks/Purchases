package handler

import (
	"net/http"

	"github.com/Alex-Blacks/Purchases/internal/logging"
	"github.com/Alex-Blacks/Purchases/internal/service"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/dto"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/helpers"
)

func LoginHandler(svc *service.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logging.LoggerFromContext(r.Context())

		var req dto.LoginRequest

		if err := helpers.DecodeJSON(w, r, logger, &req); err != nil {
			helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
			return
		}

		if err := req.Validate(); err != nil {
			helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
			return
		}

		token, err := svc.Login(r.Context(), req.Email, req.Password)
		if err != nil {
			helpers.WriteDomainError(w, logger, err, map[string]any{"email": req.Email})
			return
		}

		resp := dto.LoginResponse{Token: token}
		helpers.WriteJSON(w, logger, http.StatusOK, resp)
	}
}
