package physicalauth

import (
	"time"
)

func Dummy() (bool, error) {
	time.Sleep(2 * time.Second)

	return true, nil
}
