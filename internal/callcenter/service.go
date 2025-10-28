package callcenter

import (
	"call-center-api/models"
	"call-center-api/pkg/database"
	"context"
)

type CallCenterService interface {
	PublishCall(ctx context.Context, call models.IncomingCall) error
}

type callCenterService struct {
	kafka *database.KafkaProducer
}

func NewCallCenterService(kafka *database.KafkaProducer) CallCenterService {
	return &callCenterService{kafka: kafka}
}

func (s *callCenterService) PublishCall(ctx context.Context, call models.IncomingCall) error {
	return s.kafka.PublishIncomingCall(ctx, call)
}
