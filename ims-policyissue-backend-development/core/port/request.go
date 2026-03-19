package port

// MetadataRequest for pagination in list endpoints
type MetadataRequest struct {
	Skip     uint64 `form:"skip,default=0" validate:"omitempty"`
	Limit    uint64 `form:"limit,default=10" validate:"omitempty"`
	OrderBy  string `form:"orderBy" validate:"omitempty"`
	SortType string `form:"sortType" validate:"omitempty"`
}
// SetDefaults sets default values for pagination
// SetDefaults sets default values for pagination
func (m *MetadataRequest) SetDefaults() {
    if m.Limit == 0 {
        m.Limit = 20
    }
}

// Offset calculates the offset for database queries
func (m *MetadataRequest) Offset() int {
    return int(m.Skip)
}