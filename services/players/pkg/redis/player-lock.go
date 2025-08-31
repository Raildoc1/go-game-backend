package redis

import "fmt"

func PlayerLockKey(userID int64) string {
	return fmt.Sprintf("lock:player:%v", userID)
}
