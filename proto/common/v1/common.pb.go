// Code generated manually. DO NOT EDIT.
// Run `make proto-gen` to regenerate from proto/common/v1/common.proto

package commonv1

// Empty is the canonical empty message.
type Empty struct{}

// PageRequest carries pagination parameters.
type PageRequest struct {
	Page     int32 `json:"page"`
	PageSize int32 `json:"page_size"`
}

// PageResponse carries pagination metadata.
type PageResponse struct {
	Total    int64 `json:"total"`
	Page     int32 `json:"page"`
	PageSize int32 `json:"page_size"`
}
