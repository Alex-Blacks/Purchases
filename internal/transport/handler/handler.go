package handler

type Handlers struct {
	User    *UserHandler
	Store   *StoreHandler
	Product *ProductHandler
	Order   *OrderHandler
	Auth    *AuthHandler
}

func NewHandlers(
	userSvc ServiceUserInterface,
	storeSvc ServiceStoreInterface,
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
		Product: &ProductHandler{productService: productSvc},
		Order:   &OrderHandler{orderService: orderSvc},
		Auth:    &AuthHandler{authService: authSvc},
	}
}
