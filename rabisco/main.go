package rabisco

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/sirupsen/logrus"
)

var (
	ErrNotFound    = errors.New("Entity not found")
	ErrInvalidArgs = errors.New("Invalid arguments")
)

type Player struct {
	ID    string `json:"id"`
	Round int    `json:"round"`
}

type Round struct {
	Type  RoundType `json:"type"`
	Value string    `json:"value"`
	Round int       `json:"round"`
}

type Room struct {
	ID      string    `json:"id"`
	Players []Player  `json:"players"`
	Rounds  []Round   `json:"rounds"`
	State   RoomState `json:"state"`
	Round   int       `json:"round"`
}

func (r *Round) String() string {
	return fmt.Sprintf("<%s,%d>", r.Type.String(), r.Round)
}

type Backend interface {
	CreateRoom(ctx context.Context, logger logrus.FieldLogger, roomID, name string) error
	GetRoom(ctx context.Context, logger logrus.FieldLogger, roomID, playerID string) (*Room, error)
	JoinRoom(ctx context.Context, logger logrus.FieldLogger, playerID, roomID string) error
	Play(ctx context.Context, logger logrus.FieldLogger, roomID, playerID string, round *Round, notify chan bool) error
	Start(ctx context.Context, logger logrus.FieldLogger, roomID string) error
	NextScore(ctx context.Context, logger logrus.FieldLogger, roomID string) error
}

func NewBackend(ctx context.Context, logger logrus.FieldLogger, dbURL string) (Backend, error) {
	rand.Seed(time.Now().UTC().UnixNano())
	client, err := mongo.NewClient(dbURL)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Unable to start mongo client")
		return nil, err
	}
	if err := client.Connect(context.Background()); err != nil {
		logger.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Unable to connect to mongo")
		return nil, err
	}
	return &rabisco{roomsColl: client.Database("rabisco").Collection("rooms")}, nil
}

type rabisco struct {
	roomsColl *mongo.Collection
}
