package domain

import "github.com/tikiclone/tiki/packages/go-shared/pkg/errors"

var (
	ErrProductNotFound    = errors.NewNotFound("Product not found")
	ErrCategoryNotFound   = errors.NewNotFound("Category not found")
	ErrSKUNotFound        = errors.NewNotFound("SKU not found")
	ErrInvalidProductData = errors.NewValidation("Invalid product data")
	ErrDuplicateProduct   = errors.NewDuplicate("Product already exists")
	ErrInvalidCategory    = errors.NewValidation("Invalid category")
)
