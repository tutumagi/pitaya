package utils

import (
	"time"

	"github.com/tutumagi/pitaya/logger"
	"go.uber.org/zap"
)

//SecondTimeStringToTime 精确到秒的时间格式转成time
func SecondTimeStringToNanoTime(timeStr string) (int64, error) {
	startTime, err := time.ParseInLocation("2006-01-02 15:04:05", timeStr, time.Local)
	if err != nil {
		logger.Error("parse time error", zap.Error(err))
		return 0, err
	}
	return startTime.UnixNano(), nil
}
