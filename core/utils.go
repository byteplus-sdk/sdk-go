package core

import (
	"log"
	"runtime/debug"
	"strings"
)

func AsyncExecute(runnable func()) {
	go func(run func()) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[ByteplusSDK] async execute occur panic, "+
					"please feedback to bytedance, err:%v trace:\n%s", r, string(debug.Stack()))
			}
		}()
		run()
	}(runnable)
}

func IsNetError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), netErrMark)
}

func IsTimeoutError(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "timeout")
}