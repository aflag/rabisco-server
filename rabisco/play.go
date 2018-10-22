package rabisco

import (
	"context"
	"fmt"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/sirupsen/logrus"
)

func validateRound(r *Round) bool {
	return round2Type(r.Round) == r.Type
}

func (r *rabisco) Play(ctx context.Context, logger logrus.FieldLogger, roomID, playerID string, round *Round, notify chan bool) error {
	logger = logger.WithFields(logrus.Fields{
		"roomId":   roomID,
		"playerId": playerID,
		"round":    round.String(),
	})
	logger.Info("Making a move")
	if !validateRound(round) {
		logger.Warn("Invalid round")
		return ErrInvalidArgs
	}

	filter := map[string]interface{}{
		"_id":   roomID,
		"state": Running.String(),
		"round": round.Round,
	}
	doc := bson.NewDocument()
	if result := r.roomsColl.FindOne(ctx, filter); result == nil {
		logger.Warn("No matches")
		return ErrNotFound
	} else if err := result.Decode(doc); err != nil {
		logger.WithFields(logrus.Fields{"error": err.Error()}).Warn("Decode error")
		return err
	}

	sidx, err := findPlayerStackIndex(playerID, int64(round.Round), doc, logger)
	if err != nil {
		logger.Warn("Failed to find stack")
		return fmt.Errorf("Failed to find stack")
	}
	// add command to next round
	addr := fmt.Sprintf("stacks.%d.%d", sidx, round.Round)
	op := map[string]interface{}{
		"$set": map[string]interface{}{
			addr: map[string]string{
				"type":  round.Type.String(),
				"value": round.Value,
			},
		},
	}

	if result, err := r.roomsColl.UpdateOne(ctx, filter, op); err != nil {
		logger.Error("Failed to play")
		return fmt.Errorf("Failed to play")
	} else if result.MatchedCount != 1 {
		logger.Error("Round has already passed")
		return ErrNotFound
	}

	go r.nextRound(roomID, logger, notify)

	return nil
}
