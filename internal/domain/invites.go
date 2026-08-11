package domain

import (
	"context"
	"time"
)

type StatusInvite string

const (
	StatusPending  StatusInvite = "pending"
	StatusRejected StatusInvite = "rejected"
	StatusAccepted StatusInvite = "accepted"
)

type InviteDetails struct {
	ID            int
	GroupID       int
	Group         string
	InviterUserID int
	InviterUser   string
	InviteeEmail  string
	Status        StatusInvite
	Token         string
	CreatedAt     time.Time
	ExpiresAt     time.Time
}

type InviteUpdate struct {
	GroupID       *int
	InviterUserID *int
	InviteeEmail  *string
	Status        *StatusInvite
	Token         *string
	CreatedAt     *time.Time
	ExpiresAt     *time.Time
}

type InviteRepository interface {
	CreateInvite(ctx context.Context, q Querier, groupID int, inviterUserID int, inviteeEmail string, token string) (InviteDetails, error)
	GetInviteByID(ctx context.Context, q Querier, inviteID int) (InviteDetails, error)
	GetInviteByToken(ctx context.Context, q Querier, token string) (InviteDetails, error)
	GetInviteByEmail(ctx context.Context, q Querier, groupID int, email string) (InviteDetails, error)
	UpdateInviteByID(ctx context.Context, q Querier, inviteID int, groupID int, updateInvite InviteUpdate) (InviteDetails, error)
	DeleteInviteByID(ctx context.Context, q Querier, inviteID int, groupID int) error
	ListInvites(ctx context.Context, q Querier, groupID int) ([]InviteDetails, error)
	ListAdminInvites(ctx context.Context, q Querier) ([]InviteDetails, error)
}
