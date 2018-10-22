package rabisco

import (
	"context"

	"github.com/sirupsen/logrus"
)

// CreateRoom will create a room in waiting for players state, if it doesn't
// exist yet.
//
// This is how a room looks like in the database:
//
// {
//   _id:   String,
//   name:  String,
//   round: int,
//   state: string,
//   playerIds: [string]
//   stacks: [
//     [{type: String, value: String, round: int}]
//   ]
// }
func (r *rabisco) CreateRoom(ctx context.Context, logger logrus.FieldLogger, roomID, name string) error {
	doc := map[string]string{"_id": roomID, "name": name, "state": Waiting.String()}
	logger = logger.WithFields(logrus.Fields{
		"roomId": roomID,
		"name":   name,
	})
	_, err := r.roomsColl.InsertOne(ctx, doc)
	if err != nil {
		logger.WithField("error", err).Warn("Creating room failed")
	} else {
		logger.Info("Room created")
	}

	return err
}
