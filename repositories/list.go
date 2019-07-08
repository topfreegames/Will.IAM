package repositories

// ListOptions is used by repo.ListOptions functions
type ListOptions struct {
	PageSize int
	Page     int
}

// Limit as used in queries
func (lo ListOptions) Limit() interface{} {
	if lo.PageSize == 0 {
		return nil
	}
	return lo.PageSize
}

// Offset as used in queries
func (lo ListOptions) Offset() int {
	if lo.PageSize == 0 {
		return 0
	}
	return lo.PageSize * lo.Page
}
