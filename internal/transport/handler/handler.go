package handler

type Handlers struct {
	User    *UserHandler
	Store   *StoreHandler
	Unit    *UnitHandler
	Product *ProductHandler
	Order   *OrderHandler
	Auth    *AuthHandler
}

func NewHandlers(
	userSvc ServiceUserInterface,
	storeSvc ServiceStoreInterface,
	unitSvc ServiceUnitInterface,
	productSvc ServiceProductInterface,
	orderSvc ServiceOrderInterface,
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
	if orderSvc == nil {
		panic("orderSvc is nil")
	}
	if authSvc == nil {
		panic("authSvc is nil")
	}
	return &Handlers{
		User:    &UserHandler{userService: userSvc},
		Store:   &StoreHandler{storeService: storeSvc},
		Unit:    &UnitHandler{unitService: unitSvc},
		Product: &ProductHandler{productService: productSvc},
		Order:   &OrderHandler{orderService: orderSvc},
		Auth:    &AuthHandler{authService: authSvc},
	}
}
