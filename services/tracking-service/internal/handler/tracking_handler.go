package handler

import (
	"io"
	"math/rand"
	"time"

	commonproto "food-delivery-platform/common/proto"
	"food-delivery-platform/services/tracking-service/internal/proto"
	"food-delivery-platform/services/tracking-service/internal/repository"
	"go.uber.org/zap"
)

type TrackingHandler struct {
	proto.UnimplementedTrackingServiceServer
	repo   *repository.TrackingRepo
	logger *zap.Logger
}

func NewTrackingHandler(repo *repository.TrackingRepo, logger *zap.Logger) *TrackingHandler {
	return &TrackingHandler{repo: repo, logger: logger}
}

func (h *TrackingHandler) TrackDelivery(req *proto.TrackDeliveryRequest, stream proto.TrackingService_TrackDeliveryServer) error {
	h.logger.Info("Track delivery", zap.String("order_id", req.OrderId))

	// Send simulated location updates
	for i := 0; i < 10; i++ {
		location := &commonproto.Location{
			Lat:       28.6139 + rand.Float64()*0.01,
			Lng:       77.2090 + rand.Float64()*0.01,
			Timestamp: time.Now().Add(time.Duration(i) * time.Second).Format(time.RFC3339),
		}

		if err := stream.Send(location); err != nil {
			h.logger.Error("Send location error", zap.Error(err))
			return err
		}

		time.Sleep(2 * time.Second)
	}

	return nil
}

func (h *TrackingHandler) SendLocation(stream proto.TrackingService_SendLocationServer) error {
	for {
		location, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&proto.TrackingResponse{Message: "Locations received"})
		}
		if err != nil {
			h.logger.Error("Receive location error", zap.Error(err))
			return err
		}

		// In a real implementation, you'd save this location to the database
		h.logger.Info("Received location",
			zap.Float64("lat", location.Lat),
			zap.Float64("lng", location.Lng),
			zap.String("timestamp", location.Timestamp))

		// For demo, just acknowledge receipt
	}
}
