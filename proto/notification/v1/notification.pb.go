// Code generated manually. DO NOT EDIT.

package notificationv1

// NotificationProto is a single notification record.
type NotificationProto struct {
	Id         string `json:"id"`
	UserId     string `json:"user_id"`
	Type       string `json:"type"`
	ActorId    string `json:"actor_id"`
	TargetType string `json:"target_type"`
	TargetId   string `json:"target_id"`
	Message    string `json:"message"`
	IsRead     bool   `json:"is_read"`
	CreatedAt  int64  `json:"created_at"`
}

// SendRequest is the request for Send.
type SendRequest struct {
	UserId     string `json:"user_id"`
	Type       string `json:"type"`
	ActorId    string `json:"actor_id"`
	TargetType string `json:"target_type"`
	TargetId   string `json:"target_id"`
	Message    string `json:"message"`
}

// SendResponse is the response from Send.
type SendResponse struct {
	Id string `json:"id"`
}

// ListRequest is the request for List.
type ListRequest struct {
	UserId   string `json:"user_id"`
	Page     int32  `json:"page"`
	PageSize int32  `json:"page_size"`
}

// ListResponse is the response from List.
type ListResponse struct {
	Notifications []*NotificationProto `json:"notifications"`
	Total         int64                `json:"total"`
}

// MarkReadRequest is the request for MarkRead.
type MarkReadRequest struct {
	UserId string   `json:"user_id"`
	Ids    []string `json:"ids"`
}

// UserRequest carries a single user ID.
type UserRequest struct {
	UserId string `json:"user_id"`
}

// CountResponse carries a count.
type CountResponse struct {
	Count int64 `json:"count"`
}
