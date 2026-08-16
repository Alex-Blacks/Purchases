package handler

import (
	"context"
	"net/http"
	"strings"

	"github.com/Alex-Blacks/Purchases/internal/authctx"
	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/Alex-Blacks/Purchases/internal/logging"
	"github.com/Alex-Blacks/Purchases/internal/policy"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/dto"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/helpers"
)

type ServiceUnitInterface interface {
	CreateUnit(ctx context.Context, actor policy.Actor, name string, shortName string) (domain.UnitDetails, error)
	ListUnits(ctx context.Context) ([]domain.UnitDetails, error)
}

type UnitHandler struct {
	unitService ServiceUnitInterface
}

func (h *UnitHandler) CreateUnitHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := authctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req dto.UnitRequest
	if err := helpers.DecodeJSON(w, r, logger, &req); err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		helpers.WriteError(w, logger, http.StatusBadRequest, "empty name")
		return
	}

	if strings.TrimSpace(req.ShortName) == "" {
		helpers.WriteError(w, logger, http.StatusBadRequest, "empty short name")
		return
	}

	unit, err := h.unitService.CreateUnit(ctx, actor, req.Name, req.ShortName)
	if err != nil {
		helpers.WriteDomainError(w, logger, err, req)
		return
	}

	resp := dto.ToUnitResponse(unit)

	helpers.WriteJSON(w, logger, http.StatusCreated, resp)
}

// ListUnitsHandler godoc
//
// @Security BearerAuth
// @Summary List units
// @Description List units
// @Tags units
// @Produce json
// @Success 200 {array} dto.UnitResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Failure 503 {object} dto.ErrorResponse
// @Router /private/units [get]
func (u *UnitHandler) ListUnitsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	list, err := u.unitService.ListUnits(ctx)
	if err != nil {
		helpers.WriteDomainError(w, logger, err, nil)
		return
	}

	resp := dto.ToUnitResponse(list)

	helpers.WriteJSON(w, logger, http.StatusOK, resp)

}
