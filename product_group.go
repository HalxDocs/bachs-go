package bachs

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// ProductGroupService provides methods for managing product groups: named
// sets of products sold together, for example behind a SELECTION-mode
// checkout where the customer picks one from the group.
type ProductGroupService struct {
	service
}

// ProductGroup is a named set of products with the resolved product records.
type ProductGroup struct {
	// ID uniquely identifies the product group (for example "pgrp_...").
	ID string `json:"id"`

	// OrganizationID is the organization that owns the group.
	OrganizationID string `json:"organization_id"`

	// Name of the group.
	Name string `json:"name"`

	// Products in the group, as full product records.
	Products []Product `json:"products"`

	// CreatedAt is when the group was created.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the group was last updated.
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateProductGroupRequest is the payload for ProductGroups.Create.
type CreateProductGroupRequest struct {
	// Name of the group. Required.
	Name string `json:"name"`

	// ProductIDs are the products in the group. Required: at least one.
	ProductIDs []string `json:"product_ids"`
}

// UpdateProductGroupRequest is the payload for ProductGroups.Update. Only
// the fields you send are changed.
type UpdateProductGroupRequest struct {
	// Name replaces the group's name.
	Name *string `json:"name,omitempty"`

	// ProductIDs replaces the products in the group.
	ProductIDs []string `json:"product_ids,omitempty"`
}

// Create creates a product group from catalog products.
func (s *ProductGroupService) Create(ctx context.Context, req CreateProductGroupRequest, opts ...RequestOption) (*ProductGroup, *ResponseMeta, error) {
	var out ProductGroup
	meta, err := s.request(ctx, http.MethodPost, "/product-groups", req, &out, opts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// List returns a paginated list of product groups, most recent first.
func (s *ProductGroupService) List(ctx context.Context, params ListParams) (*Page[ProductGroup], *ResponseMeta, error) {
	var env pageEnvelope[ProductGroup]
	meta, err := s.request(ctx, http.MethodGet, queryPath("/product-groups", params), nil, &env)
	if err != nil {
		return nil, meta, err
	}
	return env.page(), meta, nil
}

// Get retrieves a single product group by its ID.
func (s *ProductGroupService) Get(ctx context.Context, groupID string) (*ProductGroup, *ResponseMeta, error) {
	var out ProductGroup
	meta, err := s.request(ctx, http.MethodGet, "/product-groups/"+url.PathEscape(groupID), nil, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// Update changes a product group's name and/or products. Only the fields you
// send are changed.
func (s *ProductGroupService) Update(ctx context.Context, groupID string, req UpdateProductGroupRequest) (*ProductGroup, *ResponseMeta, error) {
	var out ProductGroup
	meta, err := s.request(ctx, http.MethodPatch, "/product-groups/"+url.PathEscape(groupID), req, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// Delete removes a product group. The products in it are unaffected.
func (s *ProductGroupService) Delete(ctx context.Context, groupID string) (*ResponseMeta, error) {
	meta, err := s.request(ctx, http.MethodDelete, "/product-groups/"+url.PathEscape(groupID), nil, nil)
	if err != nil {
		return meta, err
	}
	return meta, nil
}
