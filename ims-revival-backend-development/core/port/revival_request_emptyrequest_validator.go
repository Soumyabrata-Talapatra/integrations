package port

// Validate implements the Validator interface for EmptyRequest
func (t *EmptyRequest) Validate() error {
	// EmptyRequest has no fields to validate
	return nil
}
