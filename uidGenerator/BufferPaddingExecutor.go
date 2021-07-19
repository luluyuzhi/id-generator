package uidGenerator

/*
#cgo CFLAGS: -I../core/UidGenerator/include
#cgo LDFLAGS: -Llib -luidgenerator
#include <RingBuffer.h>
*/
import "C"

import (
	"fmt"
	"sync/atomic"
)

type UidGenerator interface {
	getUid() int64
	parseUID(int64)
}

const DEFAULT_SCHEDULE_INTERVAL int64 = 5 * 60 // 5 minutes
const WORKER_NAME string = "RingBuffer-Padding-Worker"
const SCHEDULE_NAME string = "RingBuffer-Padding-SCHEDULE"

// lambda supported
type BufferedUidProvider interface {
	/**
	 * Provides UID in one second
	 *
	 * @param momentInSecond
	 * @return
	 */
	provide(momentInSecond int64) []int64
}

type BufferPaddingExecutor struct {
	running             int64
	lastSecond          int64
	ringbuffer          C.struct_RingBuffer
	bufferedUidProvider BufferedUidProvider

	// const
	scheduleInterval int64 // DEFAULT_SCHEDULE_INTERVAL
}

func (bufferPaddingExecutor BufferPaddingExecutor) start() {

}

func (bufferPaddingExecutor BufferPaddingExecutor) isRunning() bool {

	return atomic.LoadInt64(&bufferPaddingExecutor.running) == 0
}

func (bufferPaddingExecutor BufferPaddingExecutor) asyncPadding() {

}

func (bufferPaddingExecutor BufferPaddingExecutor) paddingBuffer() {
	if atomic.CompareAndSwapInt64(&bufferPaddingExecutor.running, 0, 1) {
		fmt.Printf("Padding buffer is still running. {}", bufferPaddingExecutor.ringbuffer)
		return
	}

	var isFullRingBuffer = false

	for {
		if isFullRingBuffer {
			break
		}

		var uidList []int64 = bufferPaddingExecutor.bufferedUidProvider.provide(int64(atomic.AddInt64(&bufferPaddingExecutor.lastSecond, 1)))
		for _, uid := range uidList {
			isFullRingBuffer = C.put(&bufferPaddingExecutor.ringbuffer, _Ctype_long(uid)) == 0
			if isFullRingBuffer {
				break
			}
		}
		atomic.CompareAndSwapInt64(&bufferPaddingExecutor.running, 0, 1)

		fmt.Printf("End to padding buffer lastSecond:{}. {}", atomic.LoadInt64(&bufferPaddingExecutor.lastSecond), bufferPaddingExecutor.ringbuffer)
	}
}

func (bufferPaddingExecutor BufferPaddingExecutor) setScheduleInterval(scheduleInterval int64) {

	bufferPaddingExecutor.scheduleInterval = scheduleInterval
}

func (bufferPaddingExecutor BufferPaddingExecutor) shutdown() {

}
