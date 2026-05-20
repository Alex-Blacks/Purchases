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

		if !helpers.DecodeJSONHelper(w, r, logger, &req) {
			return
		}

		if err := req.Validate(); err != nil {
			logger.Warn("validation failed", "error", err)
			helpers.RespondJSON(w, http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			}, logger)
			return
		}

		token, err := svc.Login(r.Context(), req.Email, req.Password)
		if err != nil {
			logger.Warn("login failed", "email", req.Email, "error", err)
			helpers.AuthErrResponse(w, err, logger, map[string]any{
				"email": req.Email,
			})
			return
		}

		resp := dto.LoginResponse{Token: token}
		helpers.RespondJSON(w, http.StatusOK, resp, logger)
	}
}
