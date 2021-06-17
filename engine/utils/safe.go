package utils

import (
	"github.com/tutumagi/pitaya/logger"

	"go.uber.org/zap"
)

// CatchPanic call a `f` and return the err if `f` paniced
func CatchPanic(f func()) (err interface{}) {
	defer func() {
		err = recover()
		if err != nil {
			logger.Errorf("catch panic error", zap.Any("func", f), zap.Any("error", err))
		}
	}()
	f()
	return
}

// RunPanicless run `f`, return true if there is no panic
func RunPanicless(f func()) (success bool) {
	defer func() {
		err := recover()
		if err != nil {
			success = false
			logger.Errorf("catch panic error", zap.Any("func", f), zap.Any("error", err))
		} else {
			success = true
		}
	}()

	f()
	return
}

// RepeatUntilPanicless run the `f` repeatly until there is no panic
func RepeatUntilPanicless(f func()) {
	for !RunPanicless(f) {
	}
}
