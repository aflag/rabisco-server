package rabisco

import (
	"context"

	"github.com/sirupsen/logrus"
)

func (r *rabisco) JoinRoom(ctx context.Context, logger logrus.FieldLogger, playerID, roomID string) error {
	logger = logger.WithFields(logrus.Fields{
		"roomId":   roomID,
		"playerId": playerID,
	})

	filter := map[string]string{
		"_id":   roomID,
		"state": Waiting.String(),
	}
	op := map[string]interface{}{
		"$addToSet": map[string]string{
			"playerIds": playerID,
		},
	}
	result, err := r.roomsColl.UpdateOne(ctx, filter, op)
	if err != nil {
		logger.WithField("error", err).Warn("Joining room fails")
		return err
	}
	if result.MatchedCount == 0 {
		logger.Warn("Room not found")
		return ErrNotFound
	} else if result.ModifiedCount == 0 {
		logger.Info("Probably rejoining")
	}

	logger.Info("Joined succesfully")

	return nil
}
