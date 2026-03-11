// Code generated manually. DO NOT EDIT.
// Run `make proto-gen` to regenerate from proto/stats/v1/stats.proto

package statsv1

// DashboardStatsProto holds the main dashboard metrics.
type DashboardStatsProto struct {
	TotalUsers     int64   `json:"total_users"`
	NewUsersToday  int64   `json:"new_users_today"`
	TotalDownloads int64   `json:"total_downloads"`
	DownloadsToday int64   `json:"downloads_today"`
	TotalRevenue   float64 `json:"total_revenue"`
	RevenueToday   float64 `json:"revenue_today"`
	TotalGames     int64   `json:"total_games"`
	TotalOrders    int64   `json:"total_orders"`
	TotalPosts     int64   `json:"total_posts"`
	TotalReports   int64   `json:"total_reports"`
}

// ChartRequest requests chart data.
type ChartRequest struct {
	Days int32 `json:"days"`
}

// ChartResponse contains a slice of chart points.
type ChartResponse struct {
	Points []*ChartPointProto `json:"points"`
}

// ChartPointProto is a single data point.
type ChartPointProto struct {
	Date  string  `json:"date"`
	Value float64 `json:"value"`
}
