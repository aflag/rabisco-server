package rabisco

import (
	"testing"

	"github.com/sirupsen/logrus/hooks/test"
)

func TestGetRoomFirstRound(t *testing.T) {
	r, _, cleanUp := rabiscoServer()
	defer cleanUp()
	logger, _ := test.NewNullLogger()
	r.CreateRoom(nil, logger, "roomie", "rum")
	r.JoinRoom(nil, logger, "riobaldo", "roomie")
	r.JoinRoom(nil, logger, "pericles", "roomie")
	r.Start(nil, logger, "roomie")

	validateRoom := func(t *testing.T, room *Room) {
		if room.Rounds[0].Type != Description || room.Rounds[1].Type != Drawing {
			t.Log(room.Rounds[0].Type.String())
			t.Log(room.Rounds[1].Type.String())
			t.Fail()
		}
		if len(room.Rounds[0].Value) == 0 || len(room.Rounds[1].Value) != 0 {
			t.Log(room.Rounds[0].Value)
			t.Log(room.Rounds[1].Value)
			t.Fail()
		}
		if room.Round != 1 {
			t.Log(room.Round)
			t.Fail()
		}
	}
	peri, err := r.GetRoom(nil, logger, "roomie", "pericles")
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}
	validateRoom(t, peri)

	rio, err := r.GetRoom(nil, logger, "roomie", "riobaldo")
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}
	validateRoom(t, rio)

	if rio.Rounds[0].Value == peri.Rounds[0].Value {
		t.Log(rio.Rounds[0].Value)
		t.Fail()
	}
}

func TestGetRoomSecondRound(t *testing.T) {
	r, _, cleanUp := rabiscoServer()
	defer cleanUp()
	logger, _ := test.NewNullLogger()
	r.CreateRoom(nil, logger, "roomie", "rum")
	r.JoinRoom(nil, logger, "riobaldo", "roomie")
	r.JoinRoom(nil, logger, "pericles", "roomie")
	r.Start(nil, logger, "roomie")
	r.Play(nil, logger, "roomie", "riobaldo", &Round{Type: Drawing, Value: "o.O", Round: 1}, nil)
	ch := make(chan bool)
	r.Play(nil, logger, "roomie", "pericles", &Round{Type: Drawing, Value: "o.O", Round: 1}, ch)
	<-ch // wait for next turn

	validateRoom := func(t *testing.T, room *Room) {
		if room.Rounds[0].Type != Drawing || room.Rounds[1].Type != Description {
			t.Log(room.Rounds[0].Type.String())
			t.Log(room.Rounds[1].Type.String())
			t.Fail()
		}
		if room.Rounds[0].Value != "o.O" || len(room.Rounds[1].Value) != 0 {
			t.Log(room.Rounds[0].Value)
			t.Log(room.Rounds[1].Value)
			t.Fail()
		}
		if room.Round != 2 {
			t.Log(room.Round)
			t.Fail()
		}
	}
	peri, err := r.GetRoom(nil, logger, "roomie", "pericles")
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}
	validateRoom(t, peri)

	rio, err := r.GetRoom(nil, logger, "roomie", "riobaldo")
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}
	validateRoom(t, rio)
}

func TestGetRoomOnePlay(t *testing.T) {
	r, _, cleanUp := rabiscoServer()
	defer cleanUp()
	logger, _ := test.NewNullLogger()
	r.CreateRoom(nil, logger, "roomie", "rum")
	r.JoinRoom(nil, logger, "riobaldo", "roomie")
	r.JoinRoom(nil, logger, "pericles", "roomie")
	r.Start(nil, logger, "roomie")
	ch := make(chan bool)
	r.Play(nil, logger, "roomie", "pericles", &Round{Type: Drawing, Value: "o.O", Round: 1}, ch)
	if <-ch {
		t.Log("Next round")
		t.FailNow()
	}

	peri, err := r.GetRoom(nil, logger, "roomie", "pericles")
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}
	if peri.Rounds[0].Type != Description || peri.Rounds[1].Type != Drawing {
		t.Log(peri.Rounds[0].Type.String())
		t.Log(peri.Rounds[1].Type.String())
		t.Fail()
	}
	if len(peri.Rounds[0].Value) == 0 || peri.Rounds[1].Value != "o.O" {
		t.Log(peri.Rounds[0].Value)
		t.Log(peri.Rounds[1].Value)
		t.Fail()
	}
	if peri.Round != 1 {
		t.Log(peri.Round)
		t.Fail()
	}
}
