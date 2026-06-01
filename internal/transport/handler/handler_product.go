package handler

import (
	"context"
	"net/http"
	"strings"

	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/Alex-Blacks/Purchases/internal/logging"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/dto"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/helpers"
)

type ServiceProductInterface interface {
	CreateProduct(ctx context.Context, title string, unit string, categoryID int) (int, error)
	GetProduct(ctx context.Context, id int) (domain.ProductDetails, error)
	DeleteProduct(ctx context.Context, id int) error
	ListProducts(ctx context.Context) ([]domain.ProductDetails, error)

	CreateProductAlias(ctx context.Context, productID int, alias string) (int, error)
	GetProductAlias(ctx context.Context, id int) (domain.ProductAliasDetails, error)
	DeleteProductAlias(ctx context.Context, id int) error
	DeleteAllProductAliases(ctx context.Context, productID int) error
	FindProductByAlias(ctx context.Context, alias string) (string, error)
	ListProductAliases(ctx context.Context, productID int) ([]domain.ProductAliasDetails, error)
}

type ProductHandler struct {
	productService ServiceProductInterface
}

// CreateProductHandler godoc
//
// @Summary Create product
// @Description Create product
// @Tags products
// @Accept json
// @Produce json
// @Param request body dto.ProductRequest true "product payload"
// @Success 201 {object} dto.ProductCreateResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /products [post]
func (h ProductHandler) CreateProductHandler(w http.ResponseWriter, r *http.Request) {
	logger := logging.LoggerFromContext(r.Context())

	var req dto.ProductRequest

	if err := helpers.DecodeJSON(w, r, logger, &req); err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(req.Title) == "" {
		helpers.WriteError(w, logger, http.StatusBadRequest, "empty name")
		return
	}
	if strings.TrimSpace(req.Unit) == "" {
		helpers.WriteError(w, logger, http.StatusBadRequest, "empty name")
		return
	}

	if err := helpers.ValidatePositiveInt("categoryId", req.CategoryID); err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	productID, err := h.productService.CreateProduct(r.Context(), req.Title, req.Unit, req.CategoryID)
	if err != nil {
		helpers.WriteDomainError(w, logger, err, req)
		return
	}

	resp := dto.ProductCreateResponse{
		ProductID: productID,
	}
	helpers.WriteJSON(w, logger, http.StatusCreated, resp)
}

// GetProductHandler godoc
//
// @Summary Get product
// @Description Get product
// @Tags products
// @Produce json
// @Param id path int true "product ID"
// @Success 200 {object} dto.ProductResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /products/{id} [get]
func (h ProductHandler) GetProductHandler(w http.ResponseWriter, r *http.Request) {
	logger := logging.LoggerFromContext(r.Context())

	productID, err := helpers.ParsePositiveIntParam(r, "id")
	if err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	product, err := h.productService.GetProduct(r.Context(), productID)
	if err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{"productId": productID})
		return
	}

	resp := dto.ToProductResponse(product)
	helpers.WriteJSON(w, logger, http.StatusOK, resp)

}

// DeleteProductHandler godoc
//
// @Summary Delete product
// @Description Delete product
// @Tags products
// @Produce json
// @Param id path int true "product ID"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /products/{id} [delete]
func (h ProductHandler) DeleteProductHandler(w http.ResponseWriter, r *http.Request) {
	logger := logging.LoggerFromContext(r.Context())

	productID, err := helpers.ParsePositiveIntParam(r, "id")
	if err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.productService.DeleteProduct(r.Context(), productID); err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{"productId": productID})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListProductsHandler godoc
//
// @Summary List product
// @Description List product
// @Tags products
// @Produce json
// @Success 200 {array} dto.ProductResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /products [get]
func (h ProductHandler) ListProductsHandler(w http.ResponseWriter, r *http.Request) {
	logger := logging.LoggerFromContext(r.Context())

	products, err := h.productService.ListProducts(r.Context())
	if err != nil {
		helpers.WriteDomainError(w, logger, err, nil)
		return
	}

	resp := dto.ToProductsResponse(products)
	helpers.WriteJSON(w, logger, http.StatusOK, resp)
}

// CreateProductAliasHandler godoc
//
// @Summary Create product alias
// @Description Create product alias
// @Tags products
// @Accept json
// @Produce json
// @Param productId path int true "product ID"
// @Param request body dto.ProductAliasRequest true "product alias payload"
// @Success 201 {object} dto.ProductAliasCreateResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /products/{productId}/aliases [post]
func (h ProductHandler) CreateProductAliasHandler(w http.ResponseWriter, r *http.Request) {
	logger := logging.LoggerFromContext(r.Context())

	productID, err := helpers.ParsePositiveIntParam(r, "productId")
	if err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	var req dto.ProductAliasRequest

	if err := helpers.DecodeJSON(w, r, logger, &req); err != nil {
		return
	}

	if strings.TrimSpace(req.Alias) == "" {
		helpers.WriteError(w, logger, http.StatusBadRequest, "empty name")
		return
	}

	aliasID, err := h.productService.CreateProductAlias(r.Context(), productID, req.Alias)
	if err != nil {
		helpers.WriteDomainError(w, logger, err, req)
		return
	}

	resp := dto.ProductAliasCreateResponse{
		AliasID: aliasID,
	}

	helpers.WriteJSON(w, logger, http.StatusCreated, resp)
}

// GetProductAliasHandler godoc
//
// @Summary Get product alias
// @Description Get product alias
// @Tags products
// @Produce json
// @Param productId path int true "product ID"
// @Param id path int true "alias ID"
// @Success 200 {object} dto.ProductAliasResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /products/{productId}/aliases/{id} [get]
func (h ProductHandler) GetProductAliasHandler(w http.ResponseWriter, r *http.Request) {
	logger := logging.LoggerFromContext(r.Context())

	aliasID, err := helpers.ParsePositiveIntParam(r, "id")
	if err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	alias, err := h.productService.GetProductAlias(r.Context(), aliasID)
	if err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{"aliasId": aliasID})
		return
	}

	resp := dto.ToProductAliasResponse(alias)

	helpers.WriteJSON(w, logger, http.StatusOK, resp)
}

// DeleteProductAliasHandler godoc
//
// @Summary Delete product alias
// @Description Delete product alias
// @Tags products
// @Produce json
// @Param productId path int true "product ID"
// @Param id path int true "alias ID"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /products/{productId}/aliases/{id} [delete]
func (h ProductHandler) DeleteProductAliasHandler(w http.ResponseWriter, r *http.Request) {
	logger := logging.LoggerFromContext(r.Context())

	aliasID, err := helpers.ParsePositiveIntParam(r, "id")
	if err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.productService.DeleteProductAlias(r.Context(), aliasID); err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{"aliasId": aliasID})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListProductAliasesHandler godoc
//
// @Summary List product aliases
// @Description List product aliases
// @Tags products
// @Produce json
// @Param productId path int true "product ID"
// @Success 200 {array} dto.ProductAliasResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /products/{productId}/aliases [get]
func (h ProductHandler) ListProductAliasesHandler(w http.ResponseWriter, r *http.Request) {
	logger := logging.LoggerFromContext(r.Context())

	productID, err := helpers.ParsePositiveIntParam(r, "productId")
	if err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}
	aliases, err := h.productService.ListProductAliases(r.Context(), productID)
	if err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{"productId": productID})
		return
	}

	resp := dto.ToProductAliasesResponse(aliases)

	helpers.WriteJSON(w, logger, http.StatusOK, resp)
}

// DeleteAllProductAliasesHandler godoc
//
// @Summary Delete all product aliases
// @Description Delete all product aliases
// @Tags products
// @Produce json
// @Param productId path int true "product ID"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /products/{productId}/aliases [delete]
func (h ProductHandler) DeleteAllProductAliasesHandler(w http.ResponseWriter, r *http.Request) {
	logger := logging.LoggerFromContext(r.Context())

	productID, err := helpers.ParsePositiveIntParam(r, "productId")
	if err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.productService.DeleteAllProductAliases(r.Context(), productID); err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{"productId": productID})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// FindProductByAliasHandler godoc
//
// @Summary Find product by alias
// @Description Find product by alias
// @Tags products
// @Accept json
// @Produce json
// @Param alias query string true "alias"
// @Success 200 {object} dto.ProductFindResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /products/by-alias [get]
func (h ProductHandler) FindProductByAliasHandler(w http.ResponseWriter, r *http.Request) {
	logger := logging.LoggerFromContext(r.Context())

	alias := r.URL.Query().Get("alias")

	if strings.TrimSpace(alias) == "" {
		helpers.WriteError(w, logger, http.StatusBadRequest, "empty name")
		return
	}

	product, err := h.productService.FindProductByAlias(r.Context(), alias)
	if err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{"alias": alias})
		return
	}

	resp := dto.ProductFindResponse{
		Product: product,
	}
	helpers.WriteJSON(w, logger, http.StatusOK, resp)
}
