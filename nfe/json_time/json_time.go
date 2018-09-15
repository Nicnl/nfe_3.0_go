package json_time

import (
	"fmt"
	"time"
)

type JsonTime time.Time

func (t JsonTime) MarshalJSON() ([]byte, error) {
	if time.Time(t).Unix() == 0 {
		return []byte("0"), nil
	}

	return []byte(fmt.Sprintf("%d", time.Time(t).Unix())), nil
}

type JsonCurrentUnixTime time.Time

func (t JsonCurrentUnixTime) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%d", time.Now().Unix())), nil
}
