package dto

import "github.com/Alex-Blacks/Purchases/internal/domain"

type GroupRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=50"`
	AdminUserID int    `json:"adminUserId" validate:"required,gt=0"`
}

type GroupUpdateRequest struct {
	Name        *string `json:"name,omitempty" validate:"min=1,max=50"`
	AdminUserID *int    `json:"adminUserId,omitempty" validate:"gt=0"`
}

type GroupResponse struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	AdminUserID int    `json:"adminUserId"`
	AdminUser   string `json:"adminUser"`
}

func ToGroupResponse(group domain.GroupDetails) GroupResponse {
	return GroupResponse{
		ID:          group.ID,
		Name:        group.Name,
		AdminUserID: group.AdminUserID,
		AdminUser:   group.AdminUser,
	}
}

func ToGroupUpdateRequest(group GroupUpdateRequest) domain.GroupUpdate {
	return domain.GroupUpdate{
		Name:        group.Name,
		AdminUserID: group.AdminUserID,
	}
}

func ToGroupListResponse(groups []domain.GroupDetails) []GroupResponse {
	resp := make([]GroupResponse, len(groups))

	for i, g := range groups {
		resp[i] = GroupResponse{
			ID:          g.ID,
			Name:        g.Name,
			AdminUserID: g.AdminUserID,
			AdminUser:   g.AdminUser,
		}
	}

	return resp
}
