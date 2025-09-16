package errors

import "errors"

var (
	ErrorCategoryNotFoundToUpdate     = errors.New("category not found to update")
	ErrorCategoryTitleAlreadyExist    = errors.New("category title already exists")
	ErrorCategoryNotFoundToDelete     = errors.New("category not found to delete")
	ErrorCategoryHasAssociatedContent = errors.New("category has associated content")
)
