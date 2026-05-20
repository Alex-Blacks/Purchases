package handler

import (
	"net/http"

	"github.com/Alex-Blacks/Purchases/internal/logging"
	"github.com/Alex-Blacks/Purchases/internal/service/auth"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/dto"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/helpers"
)

func LoginHandler(svc *auth.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logging.LoggerFromContext(r.Context())

		var req dto.LoginRequest

		if !helpers.DecodeJSONHelper(w, r, logger, &req) {
			return
		}

		if err := helpers.ValidateStruct(req); err != nil {
			logger.Warn("invalid parameter", "error", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		token, err := svc.Login(r.Context(), req.Email, req.Password)
		if err != nil {
			helpers.DomainErrResponse(w, err, logger, map[string]any{"email": req.Email})
			return
		}

		resp := dto.LoginResponse{Token: token}
		helpers.RespondJSON(w, http.StatusOK, resp, logger)
	}
}
