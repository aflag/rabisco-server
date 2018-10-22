package rabisco

import (
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/sirupsen/logrus"
)

func notify(v bool, ch chan bool) {
	if ch != nil {
		ch <- v
	}
}

func (r *rabisco) nextRound(roomID string, logger logrus.FieldLogger, ch chan bool) {
	filterState := Running.String()
	logger = logger.WithFields(logrus.Fields{"roomId": roomID})
	defer func() {
		if r := recover(); r != nil {
			// hopefully the panic doesn't come from a bad
			// logger.
			logger.Error("Next round attempt failed drastically")
			notify(false, ch)
		}
	}()

	result := r.roomsColl.FindOne(nil, map[string]string{"_id": roomID, "state": filterState})
	doc := bson.NewDocument()
	if err := result.Decode(doc); err != nil {
		if err == mongo.ErrNoDocuments {
			logger.Debug("Not ready yet")
		} else {
			logger.WithField("error", err.Error()).Warn("Decode error")
		}
		notify(false, ch)
		return
	}

	round := doc.Lookup("round").Int64()
	logger = logger.WithFields(logrus.Fields{"round": round})

	lengths := getSubArrayLengths(doc.Lookup("stacks").MutableArray())
	if len(lengths) == 0 {
		logger.Warn("Empty stacks")
		notify(false, ch)
		return
	}
	commonLen := lengths[0]
	logger = logger.WithFields(logrus.Fields{"commonLen": commonLen})
	for _, length := range lengths {
		if length != commonLen {
			logger.Debug("Length not common")
			notify(false, ch)
			return
		}
	}

	if round < commonLen {
		logger.Info("Time for the next round")

		// operation definition
		op := map[string]interface{}{
			"$set": map[string]interface{}{
				"round": round + 1,
			},
		}
		totalPlayers := doc.Lookup("playerIds").MutableArray().Len()
		if round+1 > int64(totalPlayers) {
			logger.WithFields(logrus.Fields{"totalPlayers": totalPlayers}).Info("The next round will be scoring")
			op["$set"].(map[string]interface{})["state"] = Scoring.String()
		}

		// filter definition
		filter := map[string]interface{}{"_id": roomID, "state": filterState, "round": round}

		result, err := r.roomsColl.UpdateOne(nil, filter, op)

		if err != nil || result.ModifiedCount == 0 {
			logger.WithFields(logrus.Fields{"error": err}).Info("Next round already started")
			notify(false, ch)
		} else {
			logger.Info("Document updated")
			notify(true, ch)
		}
	} else {
		logger.Debug("Not ready yet")
		notify(false, ch)
	}
}

func getSubArrayLengths(array *bson.Array) []int64 {
	result := []int64{}
	// iterator never errs
	it, _ := array.Iterator()
	for it.Next() {
		// this can cause panic if our document is corrupted
		result = append(result, int64(it.Value().MutableArray().Len()))
	}
	return result
}
