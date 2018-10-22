package rabisco

import (
	"context"
	"fmt"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/sirupsen/logrus"
)

func (r *rabisco) GetRoom(ctx context.Context, logger logrus.FieldLogger, roomID, playerID string) (*Room, error) {
	// Take the oportunity to trigger the next round job
	go r.nextRound(roomID, logger, nil)

	logger = logger.WithFields(logrus.Fields{
		"roomId":   roomID,
		"playerId": playerID,
	})

	result := r.roomsColl.FindOne(ctx, map[string]string{"_id": roomID})
	doc := bson.NewDocument()
	if err := result.Decode(doc); err != nil {
		if err == mongo.ErrNoDocuments {
			logger.Info("No matching room")
			return nil, ErrNotFound
		} else {
			logger.WithField("error", err.Error()).Warn("Decode error")
			return nil, err
		}
	}

	room := &Room{ID: roomID}

	if arr, ok := doc.Lookup("playerIds").MutableArrayOK(); !ok {
		logger.Info("Player array not yet initialized")
	} else {
		it, err := arr.Iterator()
		if err != nil {
			logger.WithField("error", err.Error()).Warn("Invalid player array")
		} else {
			for it.Next() {
				room.Players = append(room.Players, Player{ID: it.Value().StringValue()})
			}
		}
	}

	stateRaw, ok := doc.Lookup("state").StringValueOK()
	if !ok {
		logger.Error("Invalid document")
		return nil, fmt.Errorf("Invalid document: missing <state> field")
	}
	logger = logger.WithField("state", stateRaw)

	if err := room.State.UnmarshalJSON([]byte(stateRaw)); err != nil {
		logger.WithFields(logrus.Fields{
			"error": err.Error(),
			"state": string(stateRaw),
		}).Error("Unmarshal state")
		return nil, err
	}

	if room.State == preSetup || room.State == Waiting {
		// hide the preSetup state
		room.State = Waiting
		return room, nil
	}

	round, ok := doc.Lookup("round").Int64OK()
	if !ok {
		logger.Error("Get round field")
		return nil, fmt.Errorf("Invalid document: no <round> field")
	}

	room.Round = int(round)
	logger = logger.WithField("round", room.Round)

	if room.State == Running {
		stack, err := getPlayerStack(logger, doc, playerID, round)
		if err != nil {
			logger.WithField("error", err.Error()).Info("Player stack not found")
			return nil, ErrNotFound
		}

		if len(stack) < room.Round-1 {
			logger.WithField("stackLen", len(stack)).Warn("Stack too small")
			return nil, fmt.Errorf("Invalid state: stack is at least 2 rounds behind")
		}

		top := stack[len(stack)-1]
		if top.Round < room.Round || len(stack) == 1 {
			stack = []Round{
				top,
				Round{Round: room.Round, Type: round2Type(room.Round), Value: ""},
			}
		} else {
			stack = stack[len(stack)-2:]
		}

		room.Rounds = stack
	} else {
		rounds, err := lookupStackByIndex(uint(room.Round%len(room.Players)), doc, logger)
		if err != nil {
			logger.WithField("error", err.Error()).Error("Unable to get stack")
			return nil, fmt.Errorf("Invalid state: stack is at least 2 rounds behind")
		}
		room.Rounds = rounds
	}

	return room, nil
}

func getPlayerStack(logger logrus.FieldLogger, doc *bson.Document, playerID string, round int64) ([]Round, error) {
	index, err := findPlayerStackIndex(playerID, round, doc, logger)
	if err != nil {
		return nil, err
	}
	stack, err := lookupStackByIndex(index, doc, logger)
	if err != nil {
		return nil, err
	}
	return stack, nil
}
