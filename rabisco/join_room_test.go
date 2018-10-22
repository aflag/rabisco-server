package rabisco

import (
	"testing"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/sirupsen/logrus/hooks/test"
)

func TestJoinRoom(t *testing.T) {
	r, mongoURL, cleanUp := rabiscoServer()
	defer cleanUp()
	coll := getRoomsColl(mongoURL)
	logger, _ := test.NewNullLogger()
	r.CreateRoom(nil, logger, "roomie", "rum")

	if err := r.JoinRoom(nil, logger, "vovó", "roomie"); err != nil {
		t.FailNow()
	}

	result := coll.FindOne(nil, map[string]string{"_id": "roomie"})
	doc := bson.NewDocument()
	if err := result.Decode(&doc); err != nil {
		t.Log(err.Error())
		t.FailNow()
	}
	players := doc.Lookup("playerIds").MutableArray()
	if plen := players.Len(); plen != 1 {
		t.Logf("%d", plen)
		t.Fail()
	}
	if player, err := players.Lookup(0); err != nil {
		t.Log(err.Error())
		t.Fail()
	} else if id := player.StringValue(); id != "vovó" {
		t.Log(id)
		t.Fail()
	}
}

func TestJoinRoomTwiceSingleSubdoc(t *testing.T) {
	r, mongoURL, cleanUp := rabiscoServer()
	defer cleanUp()
	coll := getRoomsColl(mongoURL)
	logger, _ := test.NewNullLogger()
	r.CreateRoom(nil, logger, "roomie", "rum")

	r.JoinRoom(nil, logger, "vovó", "roomie")
	if err := r.JoinRoom(nil, logger, "vovó", "roomie"); err != nil {
		t.Log(err.Error())
		t.Fail()
	}

	result := coll.FindOne(nil, map[string]string{"_id": "roomie"})
	doc := bson.NewDocument()
	if err := result.Decode(&doc); err != nil {
		t.Log(err.Error())
		t.FailNow()
	}
	players := doc.Lookup("playerIds").MutableArray()
	if plen := players.Len(); plen != 1 {
		t.Logf("%d", plen)
		t.Fail()
	}
	if player, err := players.Lookup(0); err != nil {
		t.Log(err.Error())
		t.Fail()
	} else if id := player.StringValue(); id != "vovó" {
		t.Log(id)
		t.Fail()
	}
}
