package rabisco

import (
	"testing"

	"github.com/sirupsen/logrus/hooks/test"
)

func TestCreateRoom(t *testing.T) {
	r, mongoURL, cleanUp := rabiscoServer()
	defer cleanUp()
	coll := getRoomsColl(mongoURL)
	logger, _ := test.NewNullLogger()

	if err := r.CreateRoom(nil, logger, "myid", "myname"); err != nil {
		t.Log(err.Error())
		t.FailNow()
	}

	result := coll.FindOne(nil, map[string]string{"_id": "myid"})
	doc := map[string]interface{}{}
	if err := result.Decode(&doc); err != nil {
		t.Logf("Decode error: %s", err.Error())
		t.FailNow()
	}
	if value := doc["_id"].(string); value != "myid" {
		t.Log(value)
		t.Fail()
	}
	if value := doc["name"].(string); value != "myname" {
		t.Log(value)
		t.Fail()
	}
	if value := doc["state"].(string); value != "waiting" {
		t.Log(value)
		t.Fail()
	}
}

func TestCreateRoomAlreadyCreated(t *testing.T) {
	r, _, cleanUp := rabiscoServer()
	defer cleanUp()
	logger, _ := test.NewNullLogger()

	r.CreateRoom(nil, logger, "myid", "myname")
	if err := r.CreateRoom(nil, logger, "myid", "myname"); err == nil {
		t.Fail()
	}
}
