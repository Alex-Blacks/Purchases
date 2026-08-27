package handler

import (
	"context"
	"net/http"

	"github.com/Alex-Blacks/Purchases/internal/actorctx"
	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/Alex-Blacks/Purchases/internal/logging"
	"github.com/Alex-Blacks/Purchases/internal/policy"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/dto"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/helpers"
	"github.com/go-playground/validator/v10"
)

type ServiceUnitInterface interface {
	Create(ctx context.Context, actor policy.Actor, params any, groupID *int) (domain.UnitDetails, error)
	Get(ctx context.Context, actor policy.Actor, id int) (domain.UnitDetails, error)
	Update(ctx context.Context, actor policy.Actor, id int, updates any) (domain.UnitDetails, error)
	Delete(ctx context.Context, actor policy.Actor, id int) error
	List(ctx context.Context, actor policy.Actor) ([]domain.UnitDetails, error)
	ListAll(ctx context.Context, actor policy.Actor) ([]domain.UnitDetails, error)
}

type UnitHandler struct {
	unitService ServiceUnitInterface
	validate    *validator.Validate
}

// CreateUnitHandler обрабатывает создание новой единицы измерения.
//
// @Security BearerAuth
// @Summary Create unit
// @Description Create a new unit (user's group or specified group for admin)
// @Tags units
// @Accept json
// @Produce json
// @Param request body dto.UnitRequest true "unit payload"
// @Success 201 {object} dto.UnitResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /private/units [post]
func (h *UnitHandler) CreateUnitHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение данных из контекста
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := actorctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 2. Декодирование и валидация тела запроса
	var req dto.UnitRequest
	if err := helpers.DecodeJSON(w, r, logger, h.validate, &req); err != nil {
		helpers.WriteDomainError(w, logger, err, req)
		return
	}

	params := domain.UnitCreate{Name: req.Name}
	// 3. Вызов сервиса для создания единицы измерения
	unit, err := h.unitService.Create(ctx, actor, params, req.GroupID)
	if err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{"params": params, "groupID": req.GroupID})
		return
	}

	// 4. Формирование и отправка ответа
	helpers.WriteJSON(w, logger, http.StatusCreated, dto.ToUnitResponse(unit))
}

// GetUnitHandler возвращает единицу измерения по ID.
//
// @Security BearerAuth
// @Summary Get unit
// @Description Get unit by ID
// @Tags units
// @Produce json
// @Param id path int true "unit ID"
// @Success 200 {object} dto.UnitResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /private/units/{id} [get]
func (h *UnitHandler) GetUnitHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение данных из контекста
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := actorctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 2. Извлечение и парсинг ID из пути
	unitID, err := helpers.ParsePositiveIntParam(r, "id")
	if err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	// 3. Вызов сервиса для получения единицы измерения
	unit, err := h.unitService.Get(ctx, actor, unitID)
	if err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{"unitID": unitID})
		return
	}

	// 4. Формирование и отправка ответа
	helpers.WriteJSON(w, logger, http.StatusOK, dto.ToUnitResponse(unit))
}

// UpdateUnitHandler обновляет единицу измерения по ID.
//
// @Security BearerAuth
// @Summary Update unit
// @Description Update unit name
// @Tags units
// @Produce json
// @Param id path int true "unit ID"
// @Param request body dto.UnitUpdateRequest true "unit payload"
// @Success 200 {object} dto.UnitResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /private/units/{id} [patch]
func (h *UnitHandler) UpdateUnitHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение данных из контекста
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := actorctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 2. Парсинг ID
	unitID, err := helpers.ParsePositiveIntParam(r, "id")
	if err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	// 3. Декодирование и валидация тела запроса
	var req dto.UnitUpdateRequest
	if err := helpers.DecodeJSON(w, r, logger, h.validate, req); err != nil {
		helpers.WriteDomainError(w, logger, err, req)
		return
	}

	// 4. Вызов сервиса для обновления
	unit, err := h.unitService.Update(ctx, actor, unitID, dto.ToUnitUpdateRequest(req))
	if err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{"unitID": unitID})
		return
	}

	// 5. Формирование и отправка ответа
	helpers.WriteJSON(w, logger, http.StatusOK, dto.ToUnitResponse(unit))
}

// DeleteUnitHandler удаляет единицу измерения по ID.
//
// @Security BearerAuth
// @Summary Delete unit
// @Description Delete unit by ID
// @Tags units
// @Produce json
// @Param id path int true "unit ID"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /private/units/{id} [delete]
func (h *UnitHandler) DeleteUnitHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение данных из контекста
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := actorctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 2. Парсинг ID
	unitID, err := helpers.ParsePositiveIntParam(r, "id")
	if err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	// 3. Вызов сервиса для удаления
	if err := h.unitService.Delete(ctx, actor, unitID); err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{"unitID": unitID})
		return
	}

	// 4. Успешное удаление без контента
	w.WriteHeader(http.StatusNoContent)
}

// ListUnitsHandler возвращает список единиц измерения, доступных пользователю (из его группы).
//
// @Security BearerAuth
// @Summary List user's units
// @Description Get list of units belonging to user's group
// @Tags units
// @Produce json
// @Success 200 {array} dto.UnitResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /private/units [get]
func (u *UnitHandler) ListUnitsHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение данных из контекста
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := actorctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 2. Вызов сервиса для получения списка
	list, err := u.unitService.List(ctx, actor)
	if err != nil {
		helpers.WriteDomainError(w, logger, err, nil)
		return
	}

	// 3. Преобразование и отправка ответа
	helpers.WriteJSON(w, logger, http.StatusOK, dto.ToListUnitResponse(list))

}

// ListAllUnitsHandler возвращает список всех единиц измерения (только для администраторов).
//
// @Security BearerAuth
// @Summary List all units (admin only)
// @Description Get list of all units (requires admin role)
// @Tags units
// @Produce json
// @Success 200 {array} dto.UnitResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /private/units/all [get]
func (u *UnitHandler) ListAllUnitsHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение данных из контекста
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := actorctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 2. Вызов сервиса для получения списка
	list, err := u.unitService.ListAll(ctx, actor)
	if err != nil {
		helpers.WriteDomainError(w, logger, err, nil)
		return
	}

	// 3. Преобразование и отправка ответа
	helpers.WriteJSON(w, logger, http.StatusOK, dto.ToListUnitResponse(list))

}
