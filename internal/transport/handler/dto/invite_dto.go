package dto

import (
	"time"

	"github.com/Alex-Blacks/Purchases/internal/domain"
)

type InviteRequest struct {
	InviteeEmail string `json:"inviteeEmail" validate:"required,email"`
}

type InviteTokenRequest struct {
	Token string `json:"token" validate:"required,gt=0"`
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
