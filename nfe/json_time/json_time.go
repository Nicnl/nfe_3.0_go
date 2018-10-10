package json_time

import (
	"fmt"
	"time"
)

// Struct vers bytes
func (t JsonTime) MarshalJSON() ([]byte, error) {
	if time.Time(t).Unix() == 0 {
		return []byte("0"), nil
	}

	return []byte(fmt.Sprintf("%d", time.Time(t).Unix())), nil
}

// Bytes vers struct
func (t *JsonTime) UnmarshalJSON(data []byte) error {
	timestamp, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return fmt.Errorf("[5a4bae7b] %s", err)
	}

	*t = JsonTime(time.Unix(timestamp, 0))
	return nil
}

type JsonCurrentUnixTime time.Time

func (t JsonCurrentUnixTime) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%d", time.Now().Unix())), nil
}
