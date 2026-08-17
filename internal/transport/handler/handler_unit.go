package handler

import (
	"context"
	"net/http"

	"github.com/Alex-Blacks/Purchases/internal/authctx"
	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/Alex-Blacks/Purchases/internal/logging"
	"github.com/Alex-Blacks/Purchases/internal/policy"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/dto"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/helpers"
	"github.com/go-playground/validator/v10"
)

type ServiceUnitInterface interface {
	CreateUnit(ctx context.Context, actor policy.Actor, name string, shortName string) (domain.UnitDetails, error)
	GetUnit(ctx context.Context, actor policy.Actor, unitID int) (domain.UnitDetails, error)
	UpdateUnit(ctx context.Context, actor policy.Actor, unitID int, updateUnit domain.UnitUpdate) (domain.UnitDetails, error)
	DeleteUnit(ctx context.Context, actor policy.Actor, unitID int) error
	ListUnits(ctx context.Context) ([]domain.UnitDetails, error)
}

type UnitHandler struct {
	unitService ServiceUnitInterface
	validate    *validator.Validate
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

	if err := h.validate.Struct(req); err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
	}

	unit, err := h.unitService.CreateUnit(ctx, actor, req.Name, req.ShortName)
	if err != nil {
		helpers.WriteDomainError(w, logger, err, req)
		return
	}

	helpers.WriteJSON(w, logger, http.StatusCreated, dto.ToUnitResponse(unit))
}

func (h *UnitHandler) GetUnitHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := authctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	unitID, err := helpers.ParsePositiveIntParam(r, "id")
	if err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	unit, err := h.unitService.GetUnit(ctx, actor, unitID)
	if err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{"unitID": unitID})
		return
	}

	helpers.WriteJSON(w, logger, http.StatusOK, dto.ToUnitResponse(unit))
}

func (h *UnitHandler) UpdateUnitHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := authctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	unitID, err := helpers.ParsePositiveIntParam(r, "id")
	if err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	var req dto.UnitUpdateRequest

	if err := helpers.DecodeJSON(w, r, logger, req); err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.validate.Struct(req); err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
	}

	unit, err := h.unitService.UpdateUnit(ctx, actor, unitID, dto.ToUnitUpdateRequest(req))
	if err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{"unitID": unitID})
		return
	}

	helpers.WriteJSON(w, logger, http.StatusOK, dto.ToUnitResponse(unit))
}

func (h *UnitHandler) DeleteUnitHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := authctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	unitID, err := helpers.ParsePositiveIntParam(r, "id")
	if err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.unitService.DeleteUnit(ctx, actor, unitID); err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{"unitID": unitID})
		return
	}

	w.WriteHeader(http.StatusNoContent)
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

	resp := dto.ToListUnitResponse(list)

	helpers.WriteJSON(w, logger, http.StatusOK, resp)

}
