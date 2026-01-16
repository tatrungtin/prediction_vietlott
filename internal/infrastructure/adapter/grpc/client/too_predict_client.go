package client

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/tool_predict/internal/application/port"
	"github.com/tool_predict/internal/domain/entity"
	"github.com/tool_predict/internal/infrastructure/logger"
	predictionpb "github.com/tool_predict/proto"
	"go.uber.org/zap"
)

// TooPredictClient implements the PredictionService port for gRPC communication
type TooPredictClient struct {
	client predictionpb.PredictionServiceClient
	conn   *grpc.ClientConn
	addr   string
}

// NewTooPredictClient creates a new gRPC client for too_predict
func NewTooPredictClient(addr string) (*TooPredictClient, error) {
	if addr == "" {
		return nil, fmt.Errorf("address cannot be empty")
	}

	return &TooPredictClient{
		addr: addr,
	}, nil
}

// connect establishes the gRPC connection
func (c *TooPredictClient) connect(timeout time.Duration) error {
	if c.conn != nil {
		return nil // Already connected
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, c.addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to too_predict at %s: %w", c.addr, err)
	}

	c.conn = conn
	c.client = predictionpb.NewPredictionServiceClient(conn)

	return nil
}

// SendPrediction sends an ensemble prediction to too_predict
func (c *TooPredictClient) SendPrediction(
	ctx context.Context,
	prediction *entity.EnsemblePrediction,
) error {
	// Ensure connection is established
	if err := c.connect(10 * time.Second); err != nil {
		return err
	}

	// Convert domain entity to protobuf
	req := c.convertToProto(prediction)

	// Send via gRPC
	resp, err := c.client.SendPrediction(ctx, req)
	if err != nil {
		return fmt.Errorf("gRPC SendPrediction failed: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("prediction rejected by too_predict: %s", resp.Message)
	}

	logger.Info("Prediction sent successfully to too_predict",
		zap.String("prediction_id", prediction.ID),
		zap.String("game_type", string(prediction.GameType)),
		zap.Ints("numbers", prediction.FinalNumbers.AsSlice()),
	)

	return nil
}

// GetPredictionStatus checks the status of a sent prediction
func (c *TooPredictClient) GetPredictionStatus(
	ctx context.Context,
	predictionID string,
) (*port.PredictionStatus, error) {
	// Ensure connection is established
	if err := c.connect(10 * time.Second); err != nil {
		return nil, err
	}

	req := &predictionpb.PredictionStatusRequest{
		PredictionId: predictionID,
	}

	resp, err := c.client.GetPredictionStatus(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("gRPC GetPredictionStatus failed: %w", err)
	}

	status := &port.PredictionStatus{
		ID:          resp.PredictionId,
		Status:      resp.Status,
		SentAt:      time.Unix(resp.SentAt, 0),
		ProcessedAt: time.Unix(resp.ProcessedAt, 0),
	}

	return status, nil
}

// Close closes the gRPC connection
func (c *TooPredictClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// convertToProto converts a domain entity to protobuf message
func (c *TooPredictClient) convertToProto(
	ensemble *entity.EnsemblePrediction,
) *predictionpb.EnsemblePredictionRequest {
	// Convert individual predictions
	predictions := make([]*predictionpb.IndividualPrediction, len(ensemble.Predictions))
	for i, pred := range ensemble.Predictions {
		predictions[i] = &predictionpb.IndividualPrediction{
			Id:            pred.ID,
			AlgorithmName: pred.AlgorithmName,
			Numbers:       convertIntSliceToInt32(pred.Numbers.AsSlice()),
			Confidence:    pred.Confidence,
			GeneratedAt:   pred.GeneratedAt.Unix(),
		}
	}

	// Convert algorithm stats
	stats := make([]*predictionpb.AlgorithmContribution, len(ensemble.AlgorithmStats))
	for i, stat := range ensemble.AlgorithmStats {
		stats[i] = &predictionpb.AlgorithmContribution{
			AlgorithmName: stat.AlgorithmName,
			Weight:        stat.Weight,
			MatchCount:    int32(stat.MatchCount),
			Confidence:    stat.Confidence,
		}
	}

	return &predictionpb.EnsemblePredictionRequest{
		Id:             ensemble.ID,
		GameType:       string(ensemble.GameType),
		FinalNumbers:   convertIntSliceToInt32(ensemble.FinalNumbers.AsSlice()),
		VotingStrategy: ensemble.VotingStrategy,
		GeneratedAt:    ensemble.GeneratedAt.Unix(),
		Predictions:    predictions,
		AlgorithmStats: stats,
	}
}

// convertIntSliceToInt32 converts []int to []int32
func convertIntSliceToInt32(input []int) []int32 {
	result := make([]int32, len(input))
	for i, v := range input {
		result[i] = int32(v)
	}
	return result
}

// Ensure TooPredictClient implements port.PredictionService
var _ port.PredictionService = (*TooPredictClient)(nil)
