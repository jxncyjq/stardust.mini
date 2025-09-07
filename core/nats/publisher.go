package nats

func (s *NatsConnection) Publish(subject string, data []byte) error {
	if s.useStream {
		_, err := s.js.Publish(subject, data)
		if err != nil {
			s.logger.Error("Failed to publish message with JetStream",
				zap.String("subject", subject),
				zap.ErrorInfo(err))
		}
		return err
	} else {
		err := s.conn.Publish(subject, data)
		if err != nil {
			s.logger.Error("Failed to publish message",
				zap.String("subject", subject),
				zap.ErrorInfo(err))
		}
		return err
	}
}
func (s *NatsConnection) PublishAsync(subject string, data []byte) error {

	if !s.useStream {
		return s.Publish(subject, data) // 降级到同步发布
	}

	_, err := s.js.PublishAsync(subject, data)
	if err != nil {
		s.logger.Error("Failed to publish async message",
			zap.String("subject", subject),
			zap.Error(err))
	}
	return err
}
