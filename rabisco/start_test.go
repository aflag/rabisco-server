package rabisco

import (
	"testing"

	"github.com/sirupsen/logrus/hooks/test"
)

func TestStart(t *testing.T) {
	r, mongoURL, cleanUp := rabiscoServer()
	defer cleanUp()
	coll := getRoomsColl(mongoURL)
	logger, _ := test.NewNullLogger()
	r.CreateRoom(nil, logger, "roomie", "rum")
	r.JoinRoom(nil, logger, "rainman", "roomie")

	if err := r.Start(nil, logger, "roomie"); err != nil {
		t.Log(err.Error())
		t.FailNow()
	}

	result := coll.FindOne(nil, map[string]string{"_id": "roomie"})
	doc := map[string]interface{}{}
	if err := result.Decode(&doc); err != nil {
		t.Logf("Decode error: %s", err.Error())
		t.FailNow()
	}
	if value := doc["state"].(string); value != "running" {
		t.Log(value)
		t.Fail()
	}
}

func TestStartAlreadyStarted(t *testing.T) {
	r, _, cleanUp := rabiscoServer()
	defer cleanUp()
	logger, _ := test.NewNullLogger()
	r.CreateRoom(nil, logger, "myid", "myname")
	r.Start(nil, logger, "myid")

	if err := r.Start(nil, logger, "myid"); err == nil {
		t.Log("Start twice didn't fail")
		t.FailNow()
	}
}
