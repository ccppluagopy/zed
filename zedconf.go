package zed

import (
	"time"
)

const (
	LOG_IDX = 0
	LOG_TAG = "zed"

	DB_DIAL_TIMEOUT   = time.Second * 10
	DB_DIAL_MAX_TIMES = 1000

	MAX_LOG_FILE_SIZE      = (1024 * 1024)
	LOG_FILE_SYNC_INTERNAL = 30
)
