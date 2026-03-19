package port

// MetadataRequest provides common pagination and sorting parameters
// Embed this in list/search request structs
type MetadataRequest struct {
	Skip     uint64 `form:"skip,default=0" validate:"omitempty"`
	Limit    uint64 `form:"limit,default=10" validate:"omitempty,max=100"`
	OrderBy  string `form:"orderBy" validate:"omitempty"`
	SortType string `form:"sortType" validate:"omitempty,oneof=asc desc ASC DESC"`
}
