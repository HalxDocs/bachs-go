package bachs

// stringPtr returns a pointer to s. Used to set optional string fields on
// request types.
func stringPtr(s string) *string {
	return &s
}
