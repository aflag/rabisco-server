package rabisco

import (
	"fmt"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/sirupsen/logrus"
)

func round2Type(round int) RoundType {
	if round%2 == 0 {
		return Description
	} else {
		return Drawing
	}
}

func stringIndexOf(needle string, hay []string) int {
	for i, value := range hay {
		if needle == value {
			return i
		}
	}
	return -1
}

func lookupPlayerIDs(doc *bson.Document) []string {
	result := []string{}

	playerIDs, ok := doc.Lookup("playerIds").MutableArrayOK()
	if !ok {
		return result
	}

	it, err := playerIDs.Iterator()
	if err != nil {
		return result
	}

	for it.Next() {
		result = append(result, it.Value().StringValue())
	}

	return result
}

func lookupStackByIndex(index uint, doc *bson.Document, logger logrus.FieldLogger) ([]Round, error) {
	logger = logger.WithFields(logrus.Fields{"stackIndex": index})

	stacks, ok := doc.Lookup("stacks").MutableArrayOK()
	if !ok {
		logger.Debug("Stacks field not found")
		return []Round{}, fmt.Errorf("Stacks not found")
	}

	value, err := stacks.Lookup(index)
	if err != nil {
		logger.Warn("Stack index not found")
		return []Round{}, err
	}

	it, err := value.MutableArray().Iterator()
	if err != nil {
		logger.Warn("Iterator problems")
		return []Round{}, err
	}

	rstack := []Round{}
	for i := 0; it.Next(); i++ {
		doc := it.Value().MutableDocument()
		var rtype RoundType
		err := (*RoundType).ReadString(&rtype, doc.Lookup("type").StringValue())
		if err != nil {
			return []Round{}, err
		}
		round := Round{
			Type:  rtype,
			Value: doc.Lookup("value").StringValue(),
			Round: i,
		}
		rstack = append(rstack, round)
	}

	return rstack, nil
}

func findPlayerStackIndex(playerID string, round int64, doc *bson.Document, logger logrus.FieldLogger) (uint, error) {
	logger = logger.WithFields(logrus.Fields{"playerId": playerID, "round": round})

	playerIDs := lookupPlayerIDs(doc)
	pidx := stringIndexOf(playerID, playerIDs)
	if pidx < 0 {
		logger.Debug("Player not found")
		return uint(0), fmt.Errorf("Player not found")
	}

	// current stack index for the given player
	sidx := (round + int64(pidx)) % int64(len(playerIDs))
	logger.WithFields(logrus.Fields{
		"playerStackIndex": sidx,
		"playerIndex":      pidx,
		"playerIdsLen":     len(playerIDs),
	}).Info("Stack index calculated")

	return uint(sidx), nil
}
