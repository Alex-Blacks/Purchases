package handler

type Handlers struct {
	User         *UserHandler
	Store        *StoreHandler
	Unit         *UnitHandler
	Product      *ProductHandler
	ProductAlias *ProductAliasHandler
	Order        *OrderHandler
	Group        *GroupHandler
	Invite       *InviteHandler
	Auth         *AuthHandler
}

func NewHandlers(
	userSvc ServiceUserInterface,
	storeSvc ServiceStoreInterface,
	unitSvc ServiceUnitInterface,
	productSvc ServiceProductInterface,
	productAliasSvc ServiceProductAliasInterface,
	orderSvc ServiceOrderInterface,
	groupSvc ServiceGroupInterface,
	inviteSvc ServiceInviteInterface,
	authSvc ServiceAuthInterface,
) *Handlers {
	if userSvc == nil {
		panic("userSvc is nil")
	}
	if storeSvc == nil {
		panic("storeSvc is nil")
	}
	if unitSvc == nil {
		panic("unitSvc is nil")
	}
	if productSvc == nil {
		panic("productSvc is nil")
	}
	if productAliasSvc == nil {
		panic("productAliasSvc is nil")
	}
	if orderSvc == nil {
		panic("orderSvc is nil")
	}
	if groupSvc == nil {
		panic("groupSvc is nil")
	}
	if inviteSvc == nil {
		panic("inviteSvc is nil")
	}
	if authSvc == nil {
		panic("authSvc is nil")
	}
	return &Handlers{
		User:         &UserHandler{userService: userSvc},
		Store:        &StoreHandler{storeService: storeSvc},
		Unit:         &UnitHandler{unitService: unitSvc},
		Product:      &ProductHandler{productService: productSvc},
		ProductAlias: &ProductAliasHandler{aliasService: productAliasSvc},
		Order:        &OrderHandler{orderService: orderSvc},
		Group:        &GroupHandler{groupService: groupSvc},
		Invite:       &InviteHandler{inviteService: inviteSvc},
		Auth:         &AuthHandler{authService: authSvc},
	}
}
