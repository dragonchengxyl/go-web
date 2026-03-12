package grpcclient

import (
	"context"
	"time"

	"github.com/studio/platform/internal/usecase"
	commonv1 "github.com/studio/platform/proto/common/v1"
	statsv1 "github.com/studio/platform/proto/stats/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// StatsGRPCClient implements usecase.StatsProvider via gRPC.
type StatsGRPCClient struct {
	client statsv1.StatsServiceClient
	conn   *grpc.ClientConn
}

// NewStatsGRPCClient dials the stats service and returns a client.
func NewStatsGRPCClient(addr string) (*StatsGRPCClient, error) {
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             3 * time.Second,
			PermitWithoutStream: true,
		}),
	)
	if err != nil {
		return nil, err
	}
	return &StatsGRPCClient{
		client: statsv1.NewStatsServiceClient(conn),
		conn:   conn,
	}, nil
}

// Close closes the underlying gRPC connection.
func (c *StatsGRPCClient) Close() error {
	return c.conn.Close()
}

// GetDashboardStats implements usecase.StatsProvider.
func (c *StatsGRPCClient) GetDashboardStats(ctx context.Context) (*usecase.DashboardStats, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := c.client.GetDashboardStats(ctx, &commonv1.Empty{})
	if err != nil {
		return nil, err
	}
	return &usecase.DashboardStats{
		TotalUsers:    resp.TotalUsers,
		NewUsersToday: resp.NewUsersToday,
		TotalPosts:    resp.TotalPosts,
		TotalReports:  resp.TotalReports,
	}, nil
}

// GetUserGrowthChart implements usecase.StatsProvider, consuming the server stream.
func (c *StatsGRPCClient) GetUserGrowthChart(ctx context.Context, days int) ([]usecase.ChartPoint, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	stream, err := c.client.StreamUserGrowth(ctx, &statsv1.ChartRequest{Days: int32(days)})
	if err != nil {
		return nil, err
	}

	var points []usecase.ChartPoint
	for {
		p, err := stream.Recv()
		if err != nil {
			break
		}
		points = append(points, usecase.ChartPoint{Date: p.Date, Value: p.Value})
	}
	return points, nil
}
