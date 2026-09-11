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

// GroupFilterQuery – структура для биндинга query-параметров
type GroupFilterQuery struct {
	AdminUserID *int    `form:"admin_user_id,omitempty" validate:"gt=0"`
	Name        *string `form:"name,omitempty" validate:"min=1,max=50"`
	Limit       int     `form:"limit" default:"10" validate:"required,min=1,max=100"`
	Offset      int     `form:"offset" default:"0" validate:"min=0"`
}

// ToDomainFilter преобразует dto.GroupFilterQuery в domain.GroupListFilter.
func (q GroupFilterQuery) ToDomainFilter() domain.GroupListFilter {
	return domain.GroupListFilter{
		Name:        q.Name,
		AdminUserID: q.AdminUserID,
		Limit:       q.Limit,
		Offset:      q.Offset,
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
