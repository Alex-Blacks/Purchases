package dto

import (
	"time"

	"github.com/Alex-Blacks/Purchases/internal/domain"
)

type InviteRequest struct {
	InviteeEmail string `json:"inviteeEmail" validate:"required,email"`
}

type InviteTokenRequest struct {
	Token string `json:"token" validate:"required,min=1"`
}

type InviteResponse struct {
	ID            int                 `json:"id"`
	GroupID       int                 `json:"groupId"`
	Group         string              `json:"group"`
	InviterUserID int                 `json:"inviterUserId"`
	InviterUser   string              `json:"inviterUser"`
	InviteeEmail  string              `json:"inviteeEmail"`
	Status        domain.StatusInvite `json:"status"`
	Token         string              `json:"token"`
	CreatedAt     time.Time           `json:"createdAt"`
	ExpiresAt     time.Time           `json:"expiresAt"`
}

func ToInviteResponse(invite domain.InviteDetails) InviteResponse {
	return InviteResponse{
		ID:            invite.ID,
		GroupID:       invite.GroupID,
		Group:         invite.Group,
		InviterUserID: invite.InviterUserID,
		InviterUser:   invite.InviterUser,
		InviteeEmail:  invite.InviteeEmail,
		Status:        invite.Status,
		Token:         invite.Token,
		CreatedAt:     invite.CreatedAt,
		ExpiresAt:     invite.ExpiresAt,
	}
}

// InviteFilterQuery – структура для биндинга query-параметров
type InviteFilterQuery struct {
	GroupIDs      []int                `form:"group_ids" validate:"dive,int,gt=0"`
	InviterUserID *int                 `form:"inviter_user_id,omitempty" validate:"gt=0"`
	InviteeEmail  *string              `form:"invitee_email,omitempty" validate:"min=1,max=50"`
	Status        *domain.StatusInvite `form:"status,omitempty" validate:"oneof=pending rejected accepted"`
	Token         *string              `form:"token,omitempty" validate:"min=1"`
	CreatedFrom   *time.Time           `form:"created_from" validate:"daterange=CreatedTo"`
	CreatedTo     *time.Time           `form:"created_to"`
	ExpiresFrom   *time.Time           `form:"expires_from" validate:"daterange=ExpiresTo"`
	ExpiresTo     *time.Time           `form:"expires_to"`
	Limit         int                  `form:"limit" default:"10" validate:"required,min=1,max=100"`
	Offset        int                  `form:"offset" default:"0" validate:"min=0"`
}

// ToDomainFilter преобразует dto.InviteFilterQuery в domain.InviteListFilter.
func (q InviteFilterQuery) ToDomainFilter() domain.InviteListFilter {
	return domain.InviteListFilter{
		GroupIDs:      q.GroupIDs,
		InviterUserID: q.InviterUserID,
		InviteeEmail:  q.InviteeEmail,
		Status:        q.Status,
		Token:         q.Token,
		CreatedFrom:   q.CreatedFrom,
		CreatedTo:     q.CreatedTo,
		ExpiresFrom:   q.ExpiresFrom,
		ExpiresTo:     q.ExpiresTo,
		Limit:         q.Limit,
		Offset:        q.Offset,
	}
}

func ToInviteListResponse(invites []domain.InviteDetails) []InviteResponse {
	resp := make([]InviteResponse, len(invites))

	for i, invite := range invites {
		resp[i] = InviteResponse{
			ID:            invite.ID,
			GroupID:       invite.GroupID,
			Group:         invite.Group,
			InviterUserID: invite.InviterUserID,
			InviterUser:   invite.InviterUser,
			InviteeEmail:  invite.InviteeEmail,
			Status:        invite.Status,
			Token:         invite.Token,
			CreatedAt:     invite.CreatedAt,
			ExpiresAt:     invite.ExpiresAt,
		}
	}

	return resp
}
