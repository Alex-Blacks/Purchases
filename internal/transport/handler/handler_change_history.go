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

type ServiceChangeHistory interface {
	List(ctx context.Context, actor policy.Actor, filter domain.HistoryListFilter) ([]domain.ChangeHistoryDetails, error)
	Count(ctx context.Context, actor policy.Actor, filter domain.HistoryListFilter) (int, error)
}

type HistoryHandler struct {
	historyService ServiceChangeHistory
	validate       *validator.Validate
}

// ListChangeHistoryHandler возвращает историю.
//
// @Security BearerAuth
// @Summary List user's histories
// @Description Get list of histories
// @Tags histories
// @Produce json
// @Param group_ids[] query []int false "Group IDs"
// @Param entity_type query string false "Entity type(store, unit, product, productAlias, order, orderItem)" Enums(store, unit, product, productAlias, order, orderItem)
// @Param entity_id query int false "Entity ID" minimum(1)
// @Param action query string false "Action(create, update, delete)" Enums(create, update, delete)
// @Param from query string false "From (RFC3339)"
// @Param to query string false "To (RFC3339)"
// @Param limit query int false "Limit" default(10) minimum(1) maximum(100)
// @Param offset query int false "Offset" default(0) minimum(0)
// @Success 200 {array} dto.ChangeHistoryResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /private/histories [get]
func (h *HistoryHandler) ListChangeHistoryHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение данных из контекста
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := actorctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 2. Биндим query-параметры в структуру
	var queryFilter dto.HistoryFilterQuery
	if err := helpers.FormDecoder.Decode(&queryFilter, r.URL.Query()); err != nil {
		logger.WarnContext(ctx, "invalid query parameters", "error", err)
		helpers.WriteError(w, logger, http.StatusBadRequest, "недопустимые параметры запроса")
		return
	}

	if err := h.validate.Struct(queryFilter); err != nil {
		helpers.WriteDomainError(w, logger, err, queryFilter)
		return
	}

	// 3. Вызов сервиса для получения списка
	list, err := h.historyService.List(ctx, actor, queryFilter.ToDomainFilter())
	if err != nil {
		helpers.WriteDomainError(w, logger, err, nil)
		return
	}

	// 4. Преобразование и отправка ответа
	helpers.WriteJSON(w, logger, http.StatusOK, dto.ToChangeHistoryListResponse(list))
}

// CountChangeHistoryHandler возвращает количество записей в истории.
//
// @Security BearerAuth
// @Summary Count user's histories
// @Description Get count of histories
// @Tags histories
// @Produce json
// @Param group_ids[] query []int false "Group IDs"
// @Param entity_type query string false "Entity type(store, unit, product, productAlias, order, orderItem)" Enums(store, unit, product, productAlias, order, orderItem)
// @Param entity_id query int false "Entity ID" minimum(1)
// @Param action query string false "Action(create, update, delete)" Enums(create, update, delete)
// @Param from query string false "From (RFC3339)"
// @Param to query string false "To (RFC3339)"
// @Param limit query int false "Limit" default(10) minimum(1) maximum(100)
// @Param offset query int false "Offset" default(0) minimum(0)
// @Success 200 {object} dto.CountResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /private/histories/count [get]
func (h *HistoryHandler) CountChangeHistoryHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение данных из контекста
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := actorctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 2. Биндим query-параметры в структуру
	var queryFilter dto.HistoryFilterQuery
	if err := helpers.FormDecoder.Decode(&queryFilter, r.URL.Query()); err != nil {
		logger.WarnContext(ctx, "invalid query parameters", "error", err)
		helpers.WriteError(w, logger, http.StatusBadRequest, "недопустимые параметры запроса")
		return
	}

	if err := h.validate.Struct(queryFilter); err != nil {
		helpers.WriteDomainError(w, logger, err, queryFilter)
		return
	}

	// 3. Вызов сервиса для получения количества
	count, err := h.historyService.Count(ctx, actor, queryFilter.ToDomainFilter())
	if err != nil {
		helpers.WriteDomainError(w, logger, err, nil)
		return
	}

	// 4. Преобразование и отправка ответа
	helpers.WriteJSON(w, logger, http.StatusOK, dto.CountResponse{Count: count})
}
