package rabisco

import (
	"encoding/json"
	"fmt"
	"strings"
)

type RoundType int64

const (
	Drawing RoundType = iota
	Description
)

func (r RoundType) String() string {
	return []string{"drawing", "description"}[r]
}

func (r RoundType) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.String())
}

func (r *RoundType) UnmarshalJSON(b []byte) error {
	str := strings.Trim(string(b), `"`)
	switch str {
	case "drawing":
		*r = Drawing
		return nil
	case "description":
		*r = Description
		return nil
	}
	return fmt.Errorf("Unknown RoundType: %s", str)
}

func (r *RoundType) ReadString(str string) error {
	return r.UnmarshalJSON([]byte(str))
}

type RoomState int64

const (
	Waiting RoomState = iota
	// The preSetup state is used only internally and never
	// exposed
	preSetup
	Running
	Scoring
)

func (r RoomState) String() string {
	return []string{"waiting", "preSetup", "running", "scoring"}[r]
}

func (r RoomState) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.String())
}

func (r *RoomState) UnmarshalJSON(b []byte) error {
	str := strings.Trim(string(b), `"`)
	switch str {
	case "waiting":
		*r = Waiting
	case "preSetup":
		*r = preSetup
	case "running":
		*r = Running
	case "scoring":
		*r = Scoring
	default:
		return fmt.Errorf("Invalid roomState: %s", str)
	}
	return nil
}
