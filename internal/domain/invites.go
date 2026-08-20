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
	Create(ctx context.Context, q Querier, groupID int, inviterUserID int, inviteeEmail string, token string) (InviteDetails, error)
	GetByID(ctx context.Context, q Querier, inviteID int) (InviteDetails, error)
	GetByToken(ctx context.Context, q Querier, token string) (InviteDetails, error)
	GetByEmail(ctx context.Context, q Querier, groupID int, email string) (InviteDetails, error)
	UpdateByID(ctx context.Context, q Querier, inviteID int, groupID int, updateInvite InviteUpdate) (InviteDetails, error)
	DeleteByID(ctx context.Context, q Querier, inviteID int, groupID int) error
	List(ctx context.Context, q Querier, groupID int) ([]InviteDetails, error)
	ListAll(ctx context.Context, q Querier) ([]InviteDetails, error)
}
