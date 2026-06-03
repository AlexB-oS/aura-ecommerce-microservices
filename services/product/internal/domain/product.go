// Package domain holds the Product Catalog service's core entities and errors.
// Like the user service's domain package it is free of any HTTP/SQL/JSON
// knowledge, so the business rules stay easy to test and reason about.
package domain

import (
	"errors"
	"time"
)

// Product is the central catalog entity. Money is always stored as an integer
// number of cents in the smallest currency unit — never a float — to avoid
// rounding errors.
type Product struct {
	ID          string
	Slug        string // url-friendly unique identifier, e.g. "drift-headphones"
	Name        string
	Tagline     string // short marketing line shown on cards
	Description string // longer copy for the product page
	Category    string
	PriceCents  int
	Currency    string
	Stock       int
	Badge       string // optional, e.g. "New" / "Best seller"
	Active      bool   // soft on/off switch without deleting the row
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Filter describes the optional constraints when listing products. The zero
// value lists everything (subject to the repository's default page size).
type Filter struct {
	Category    string // exact category match; "" = any
	Search      string // case-insensitive match against name/tagline; "" = any
	InStockOnly bool
	Limit       int
	Offset      int
}

// Domain errors. The transport layer maps these onto HTTP status codes.
var (
	ErrProductNotFound   = errors.New("product not found")
	ErrSlugTaken         = errors.New("a product with that slug already exists")
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrInvalidPrice      = errors.New("price must be greater than zero")
)
