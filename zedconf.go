package zed

const (
	LOG_IDX = 0
	LOG_TAG = "zed"

	DB_DIAL_TIMEOUT   = 10
	DB_DIAL_MAX_TIMES = 3

	MAX_LOG_FILE_SIZE      = (1024 * 1024)
	LOG_FILE_SYNC_INTERNAL = 30
)

const (
	LogCmd = iota
	LogFile
)
