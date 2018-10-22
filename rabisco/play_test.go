package rabisco

import (
	"testing"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
)

func TestPlay(t *testing.T) {
	r, mongoURL, cleanUp := rabiscoServer()
	defer cleanUp()
	coll := getRoomsColl(mongoURL)
	logger, _ := test.NewNullLogger()
	r.CreateRoom(nil, logger, "roomie", "rum")
	r.JoinRoom(nil, logger, "riobaldo", "roomie")
	r.Start(nil, logger, "roomie")

	err := r.Play(nil, logger, "roomie", "riobaldo", &Round{Type: Drawing, Value: "o.O", Round: 1}, nil)
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}

	result := coll.FindOne(nil, map[string]string{"_id": "roomie"})
	doc := bson.NewDocument()
	if err := result.Decode(&doc); err != nil {
		t.Log(err.Error())
		t.FailNow()
	}
	stack, err := doc.Lookup("stacks").MutableArray().Lookup(uint(0))
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}
	comm, err := stack.MutableArray().Lookup(uint(1))
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}
	t.Log(comm)
	commdoc := comm.MutableDocument()
	if typ := commdoc.Lookup("type").StringValue(); typ != Drawing.String() {
		t.Log(typ)
		t.Fail()
	}
	if value := commdoc.Lookup("value").StringValue(); value != "o.O" {
		t.Log(value)
		t.Fail()
	}
}

func TestPlayFinishRound(t *testing.T) {
	r, mongoURL, cleanUp := rabiscoServer()
	defer cleanUp()
	coll := getRoomsColl(mongoURL)
	logger, _ := test.NewNullLogger()
	r.CreateRoom(nil, logger, "roomie", "rum")
	r.JoinRoom(nil, logger, "riobaldo", "roomie")
	r.JoinRoom(nil, logger, "diadorim", "roomie")
	r.Start(nil, logger, "roomie")

	round := &Round{Type: Drawing, Value: "o.O", Round: 1}
	play(r, logger, t, false, "roomie", "riobaldo", round)
	play(r, logger, t, true, "roomie", "diadorim", round)

	round = &Round{Type: Description, Value: "wow", Round: 2}
	play(r, logger, t, false, "roomie", "riobaldo", round)
	play(r, logger, t, true, "roomie", "diadorim", round)

	result := coll.FindOne(nil, map[string]string{"_id": "roomie"})
	doc := bson.NewDocument()
	if err := result.Decode(&doc); err != nil {
		t.Log(err.Error())
		t.FailNow()
	}
	if state := doc.Lookup("state").StringValue(); state != "scoring" {
		t.Log(state)
		t.Fail()
	}
}

func play(b Backend, logger logrus.FieldLogger, t *testing.T, isOk bool, roomID, playerID string, round *Round) {
	ch := make(chan bool)
	if err := b.Play(nil, logger, roomID, playerID, round, ch); err != nil {
		t.Log(err.Error())
		t.FailNow()
	}
	if ok := <-ch; ok != isOk {
		t.Logf("%t != %t", ok, isOk)
		t.FailNow()
	}
}
