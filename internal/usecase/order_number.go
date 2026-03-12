package usecase

import (
	"fmt"
	"time"
)

func generateOrderNo() string {
	now := time.Now()
	return fmt.Sprintf("%s%08d", now.Format("20060102150405"), now.UnixNano()%100000000)
}
