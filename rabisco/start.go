package rabisco

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/sirupsen/logrus"
)

func (r *rabisco) Start(ctx context.Context, logger logrus.FieldLogger, roomID string) error {
	// room preconditions: state=waiting or state=preSetup
	logger = logger.WithFields(logrus.Fields{"roomId": roomID})

	logger.Info("Begin room PreSetup")
	doc, err := setPreSetupState(ctx, roomID, r.roomsColl, logger)
	if err != nil {
		return err
	}

	playerIDs, ok := doc.Lookup("playerIds").MutableArrayOK()
	if !ok {
		logger.Error("Failed to decode players")
		return fmt.Errorf("Failed to decode players")
	}

	logger = logger.WithFields(logrus.Fields{"totalPlayers": playerIDs.Len()})
	logger.Info("Create initial stacks")
	stacks := createStacks(playerIDs.Len())

	logger = logger.WithFields(logrus.Fields{"totalStacks": len(stacks)})
	logger.Info("Stacks created")

	return setRunningState(ctx, roomID, stacks, r.roomsColl, logger)
}

func setPreSetupState(ctx context.Context, roomID string, coll *mongo.Collection, logger logrus.FieldLogger) (*bson.Document, error) {
	filter := map[string]interface{}{
		"_id": roomID,
		"$or": [...]interface{}{
			map[string]string{"state": Waiting.String()},
			// if something goes wrong during this method's execution we can
			// try again
			map[string]string{"state": preSetup.String()},
		},
	}
	op := map[string]interface{}{
		"$set": map[string]interface{}{
			"state": preSetup.String(),
		},
	}

	result := coll.FindOneAndUpdate(ctx, filter, op)

	if result == nil {
		logger.Warn("Strange, result is nil")
		return nil, fmt.Errorf("nil result")
	}

	doc := bson.NewDocument()
	if err := result.Decode(doc); err != nil {
		logger.WithField("error", err).Info("Result decode failure")
		return nil, ErrNotFound
	}

	return doc, nil
}

func setRunningState(ctx context.Context, roomID string, stacks [][]map[string]string, coll *mongo.Collection, logger logrus.FieldLogger) error {
	filter := map[string]string{
		"_id":   roomID,
		"state": preSetup.String(),
	}
	op := map[string]interface{}{
		"$set": map[string]interface{}{
			"stacks": stacks,
			"state":  Running.String(),
			"round":  1,
		},
	}

	_, err := coll.UpdateOne(ctx, filter, op)

	if err != nil {
		logger.WithFields(logrus.Fields{"error": err}).Warn("Starting game failed")
		return err
	}

	logger.Info("Success starting game")
	return nil
}

func createStacks(totalPlayers int) (stacks [][]map[string]string) {
	perms := rand.Perm(len(seeds))

	for i := 0; i < totalPlayers; i++ {
		stack := []map[string]string{
			{"type": Description.String(), "value": seeds[perms[i]]},
		}
		stacks = append(stacks, stack)
	}

	return
}
