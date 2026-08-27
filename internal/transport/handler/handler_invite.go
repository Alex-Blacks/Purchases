package handler

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"

	"github.com/Alex-Blacks/Purchases/internal/actorctx"
	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/Alex-Blacks/Purchases/internal/logging"
	"github.com/Alex-Blacks/Purchases/internal/policy"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/dto"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/helpers"
	"github.com/go-playground/validator/v10"
)

type ServiceInviteInterface interface {
	Create(ctx context.Context, actor policy.Actor, inviteeEmail string) (domain.InviteDetails, error)
	GetByID(ctx context.Context, actor policy.Actor, inviteID int) (domain.InviteDetails, error)
	DeleteByID(ctx context.Context, actor policy.Actor, inviteID int) error
	List(ctx context.Context, actor policy.Actor) ([]domain.InviteDetails, error)
	ListAll(ctx context.Context, actor policy.Actor) ([]domain.InviteDetails, error)
	Accept(ctx context.Context, actor policy.Actor, token string) error
	Reject(ctx context.Context, actor policy.Actor, token string) error
}

type InviteHandler struct {
	inviteService ServiceInviteInterface
	validate      *validator.Validate
}

// CreateInviteHandler обрабатывает создание нового приглашения.
//
// @Security BearerAuth
// @Summary Create invite
// @Description Create a new invite
// @Tags invites
// @Accept json
// @Produce json
// @Param request body dto.InviteRequest true "invite payload"
// @Success 201 {object} dto.InviteResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /private/invites [post]
func (h *InviteHandler) CreateInviteHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение данных из контекста
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := actorctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 2. Декодирование и валидация тела запроса
	var req dto.InviteRequest
	if err := helpers.DecodeJSON(w, r, logger, h.validate, &req); err != nil {
		helpers.WriteDomainError(w, logger, err, req)
		return
	}

	// 3. Вызов сервиса
	invite, err := h.inviteService.Create(ctx, actor, req.InviteeEmail)
	if err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{"inviteEmail": fmt.Sprintf("%x", sha256.Sum256([]byte(req.InviteeEmail)))})
		return
	}

	// 4. Формирование и отправка ответа
	helpers.WriteJSON(w, logger, http.StatusCreated, dto.ToInviteResponse(invite))
}

// GetInviteHandler возвращает приглашение по ID.
//
// @Security BearerAuth
// @Summary Get invite
// @Description Get invite by ID
// @Tags invites
// @Produce json
// @Param id path int true "invite ID"
// @Success 200 {object} dto.InviteResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /private/invites/{id} [get]
func (h InviteHandler) GetInviteHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение данных из контекста
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := actorctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 2. Извлечение и парсинг ID из пути
	inviteID, err := helpers.ParsePositiveIntParam(r, "id")
	if err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	// 3. Вызов сервиса для получения приглашения
	invite, err := h.inviteService.GetByID(ctx, actor, inviteID)
	if err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{"inviteId": inviteID})
		return
	}

	// 4. Формирование и отправка ответа
	helpers.WriteJSON(w, logger, http.StatusOK, dto.ToInviteResponse(invite))
}

// DeleteInviteHandler удаляет приглашение по ID.
//
// @Security BearerAuth
// @Summary Delete invite
// @Description Delete invite by ID
// @Tags invites
// @Produce json
// @Param id path int true "invite ID"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /private/invites/{id} [delete]
func (h InviteHandler) DeleteInviteHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение данных из контекста
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := actorctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 2. Парсинг ID
	inviteID, err := helpers.ParsePositiveIntParam(r, "id")
	if err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	// 3. Вызов сервиса для удаления
	if err := h.inviteService.DeleteByID(ctx, actor, inviteID); err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{"inviteId": inviteID})
		return
	}

	// 4. Успешное удаление без контента
	w.WriteHeader(http.StatusNoContent)
}

// ListInvitesHandler возвращает список приглашений, доступных пользователю (из его группы).
//
// @Security BearerAuth
// @Summary List user's invites
// @Description Get list of invites belonging to user's group
// @Tags invites
// @Produce json
// @Success 200 {array} dto.InviteResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /private/invites [get]
func (h InviteHandler) ListInvitesHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение данных из контекста
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := actorctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 2. Вызов сервиса для получения списка
	list, err := h.inviteService.List(ctx, actor)
	if err != nil {
		helpers.WriteDomainError(w, logger, err, nil)
		return
	}

	// 3. Преобразование и отправка ответа
	helpers.WriteJSON(w, logger, http.StatusOK, dto.ToInviteListResponse(list))
}

// ListAllInvitesHandler возвращает список всех приглашений (только для администраторов).
//
// @Security BearerAuth
// @Summary List all invites
// @Description Get list of invites
// @Tags invites
// @Produce json
// @Success 200 {array} dto.InviteResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /private/invites/all [get]
func (h InviteHandler) ListAllInvitesHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение данных из контекста
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := actorctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 2. Вызов сервиса для получения списка
	list, err := h.inviteService.ListAll(ctx, actor)
	if err != nil {
		helpers.WriteDomainError(w, logger, err, nil)
		return
	}

	// 3. Преобразование и отправка ответа
	helpers.WriteJSON(w, logger, http.StatusOK, dto.ToInviteListResponse(list))
}

// AcceptInviteHandler принимает приглашение по токену.
//
// @Security BearerAuth
// @Summary Accept invite
// @Description Accept invite by token
// @Tags invites
// @Produce json
// @Param request body dto.InviteTokenRequest true "invite payload"
// @Success 200 "Ok"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /private/invites/accept [put]
func (h InviteHandler) AcceptInviteHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение данных из контекста
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := actorctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 2. Декодирование и валидация тела запроса
	var req dto.InviteTokenRequest
	if err := helpers.DecodeJSON(w, r, logger, h.validate, &req); err != nil {
		helpers.WriteDomainError(w, logger, err, req)
		return
	}

	// 3. Вызов сервиса
	if err := h.inviteService.Accept(ctx, actor, req.Token); err != nil {
		helpers.WriteDomainError(w, logger, err, nil)
		return
	}

	// 4. Успешный приём
	w.WriteHeader(http.StatusOK)
}

// RejectInviteHandler отклоняет приглашение по токену.
//
// @Security BearerAuth
// @Summary Reject invite
// @Description Reject invite by token
// @Tags invites
// @Produce json
// @Param request body dto.InviteTokenRequest true "invite payload"
// @Success 200 "Ok"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /private/invites/reject [put]
func (h InviteHandler) RejectInviteHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение данных из контекста
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := actorctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 2. Декодирование и валидация тела запроса
	var req dto.InviteTokenRequest
	if err := helpers.DecodeJSON(w, r, logger, h.validate, &req); err != nil {
		helpers.WriteDomainError(w, logger, err, req)
		return
	}

	// 3. Вызов сервиса
	if err := h.inviteService.Reject(ctx, actor, req.Token); err != nil {
		helpers.WriteDomainError(w, logger, err, nil)
		return
	}

	// 4. Успешный отказ
	w.WriteHeader(http.StatusOK)
}
