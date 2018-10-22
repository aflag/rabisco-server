package rabisco

import (
	"context"

	"github.com/sirupsen/logrus"
)

func (r *rabisco) NextScore(ctx context.Context, logger logrus.FieldLogger, roomID string) error {
	logger = logger.WithFields(logrus.Fields{
		"roomId": roomID,
	})
	filter := map[string]string{
		"_id":   roomID,
		"state": Scoring.String(),
	}
	op := map[string]interface{}{
		"$inc": map[string]int{"round": 1},
	}

	result, err := r.roomsColl.UpdateOne(ctx, filter, op)

	if err != nil {
		logger.WithField("error", err).Warn("Joining room fails")
		return err
	}
	if result.MatchedCount == 0 {
		logger.Warn("Room not found")
		return ErrNotFound
	}

	return nil
}
