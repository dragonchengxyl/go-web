// Code generated manually. DO NOT EDIT.

package moderationv1

// ReviewTextRequest is the request for ReviewText.
type ReviewTextRequest struct {
	Text string `json:"text"`
}

// ReviewImageRequest is the request for ReviewImage.
type ReviewImageRequest struct {
	Url string `json:"url"`
}

// ReviewResult is the result of a moderation review.
type ReviewResult struct {
	Decision string   `json:"decision"`
	Labels   []string `json:"labels"`
}
