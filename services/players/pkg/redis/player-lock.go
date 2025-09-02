// Package redis contains service specific Redis helpers
package redis

import "fmt"

// PlayerLockKey created key that is used for shared redis lock
// that meant to be used for executing player logic sequentially
func PlayerLockKey(userID int64) string {
	return fmt.Sprintf("lock:player:%v", userID)
}
