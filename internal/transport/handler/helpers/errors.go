package helpers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/Alex-Blacks/Purchases/internal/policy"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/dto"
	"github.com/go-playground/validator/v10"
)

// errorMapping хранит соответствие ошибок -> HTTP статус + сообщение (пока только русские).
var errorMapping = map[error]struct {
	Code int
	Msg  string
}{
	// Ошибки валидации (400)
	domain.ErrInvalidInput:             {http.StatusBadRequest, "некорректные данные запроса"},
	domain.ErrEmptyName:                {http.StatusBadRequest, "имя не может быть пустым"},
	domain.ErrNoFieldsToUpdate:         {http.StatusBadRequest, "нет полей для обновления"},
	domain.ErrInvalidGroupID:           {http.StatusBadRequest, "идентификатор группы должен быть положительным числом"},
	domain.ErrGroupIDRequired:          {http.StatusBadRequest, "для этой операции требуется идентификатор группы"},
	domain.ErrSelfInvite:               {http.StatusBadRequest, "нельзя пригласить самого себя"},
	domain.ErrInviteRejected:           {http.StatusBadRequest, "это приглашение было отклонено"},
	domain.ErrInviteExpired:            {http.StatusBadRequest, "срок действия приглашения истёк"},
	domain.ErrInviteRejectedStillValid: {http.StatusBadRequest, "это приглашение отклонено и не может быть использовано"},
	domain.ErrInvalidCredentials:       {http.StatusUnauthorized, "неверный email или пароль"},

	// Ошибки конфликтов (409)
	domain.ErrAlreadyExists:        {http.StatusConflict, "ресурс уже существует"},
	domain.ErrConflict:             {http.StatusConflict, "конфликт с существующими данными"},
	domain.ErrEmailConflict:        {http.StatusConflict, "этот email уже зарегистрирован"},
	domain.ErrUserAlreadyInGroup:   {http.StatusConflict, "пользователь уже является участником группы"},
	domain.ErrInviteAlreadyPending: {http.StatusConflict, "для этого пользователя уже есть ожидающее приглашение"},

	// Ошибки доступа/статуса (403)
	domain.ErrGroupIDNotAllowed: {http.StatusForbidden, "вам запрещено указывать идентификатор группы"},
	domain.ErrStatusBlocked:     {http.StatusForbidden, "ваш аккаунт заблокирован"},
	policy.ErrForbidden:         {http.StatusForbidden, "у вас нет прав на выполнение этого действия"},

	// Ошибка "не найдено" (404)
	domain.ErrNotFound: {http.StatusNotFound, "запрашиваемый ресурс не найден"},
}

// WriteDomainError обрабатывает ошибки:
// - доменные (из errorMapping)
// - ошибки валидации (validator.ValidationErrors)
// - контекст (таймаут/отмена)
// - всё остальное -> Internal Server Error
func WriteDomainError(w http.ResponseWriter, logger *slog.Logger, err error, req any) {
	// Проверка контекста
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		WriteError(w, logger, http.StatusServiceUnavailable, "время запроса истекло или он был отменён")
		return
	}

	// Обработка ошибок валидации (от DecodeJSON)
	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		writeValidationErrors(w, logger, validationErrs)
		return
	}

	// Поиск в маппинге доменных ошибок
	for domainErr, resp := range errorMapping {
		if errors.Is(err, domainErr) {
			WriteError(w, logger, resp.Code, resp.Msg)
			return
		}
	}

	// Неизвестная ошибка
	WriteInternalError(w, logger, err, req)
}

// writeValidationErrors формирует ответ с деталями по каждому полю
func writeValidationErrors(w http.ResponseWriter, logger *slog.Logger, errs validator.ValidationErrors) {
	details := make(map[string]string)
	for _, e := range errs {
		// Поле, тег валидации и параметр (если есть)
		field := e.Field()
		tag := e.Tag()
		param := e.Param()

		var msg string
		switch tag {
		case "required":
			msg = "поле обязательно"
		case "email":
			msg = "неверный формат email"
		case "min":
			msg = "значение должно быть не менее " + param
		case "max":
			msg = "значение должно быть не более " + param
		case "len":
			msg = "длина должна быть " + param
		case "oneof":
			msg = "допустимые значения: " + param
		default:
			msg = "некорректное значение"
		}
		details[field] = msg
	}

	logger.Warn("validation failed", "errors", details)
	WriteJSON(w, logger, http.StatusBadRequest, dto.ErrorResponse{
		Error:   "некорректные данные",
		Details: details,
	})
}
