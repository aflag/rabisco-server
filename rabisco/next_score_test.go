package rabisco

import (
	"testing"

	"github.com/sirupsen/logrus/hooks/test"
)

func NextRoundTest(t *testing.T) {
	r, _, cleanUp := rabiscoServer()
	defer cleanUp()
	logger, _ := test.NewNullLogger()
	r.CreateRoom(nil, logger, "roomie", "rum")
	r.JoinRoom(nil, logger, "riobaldo", "roomie")
	r.Start(nil, logger, "roomie")
	r.Play(nil, logger, "roomie", "riobaldo", &Round{Type: Drawing, Value: "o.O", Round: 1}, nil)

	if err := r.NextScore(nil, logger, "roomie"); err != nil {
		t.Log(err.Error())
		t.FailNow()
	}

	rio, err := r.GetRoom(nil, logger, "roomie", "riobaldo")
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}

	if rio.Round != 2 {
		t.Log(rio.Round)
		t.Fail()
	}
}

func NextRoundTwiceTest(t *testing.T) {
	r, _, cleanUp := rabiscoServer()
	defer cleanUp()
	logger, _ := test.NewNullLogger()
	r.CreateRoom(nil, logger, "roomie", "rum")
	r.JoinRoom(nil, logger, "riobaldo", "roomie")
	r.Start(nil, logger, "roomie")
	r.Play(nil, logger, "roomie", "riobaldo", &Round{Type: Drawing, Value: "o.O", Round: 1}, nil)

	if err := r.NextScore(nil, logger, "roomie"); err != nil {
		t.Log(err.Error())
		t.FailNow()
	}
	if err := r.NextScore(nil, logger, "roomie"); err != nil {
		t.Log(err.Error())
		t.FailNow()
	}

	rio, err := r.GetRoom(nil, logger, "roomie", "riobaldo")
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}

	if rio.Round != 3 {
		t.Log(rio.Round)
		t.Fail()
	}
}
