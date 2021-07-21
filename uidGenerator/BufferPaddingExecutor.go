package uidGenerator

/*
#cgo CFLAGS: -I${SRCDIR}/core/UidGenerator/include
#cgo LDFLAGS: -L${SRCDIR}/lib -luidgenerator
#include <RingBuffer.h>
*/
import "C"
import (
	"fmt"
	"sync/atomic"
	"time"
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
	taskPool            *Pool

	ticker *time.Ticker
	// const
	scheduleInterval int64 // DEFAULT_SCHEDULE_INTERVAL
}

func NewBufferPaddingExecutor(usingSchedule bool) *BufferPaddingExecutor {

	var p BufferPaddingExecutor

	atomic.StoreInt64(&p.running, 0)
	atomic.StoreInt64(&p.lastSecond, time.Now().Unix())
	p.taskPool = NewPool(4)
	// this.ringBuffer = ringBuffer;
	p.scheduleInterval = 300
	// initialize schedule thread
	if usingSchedule {
		p.ticker = time.NewTicker(time.Duration(p.scheduleInterval))
	} else {
		p.ticker = nil
	}

	return &p
}

func (bufferPaddingExecutor BufferPaddingExecutor) start() {
	if bufferPaddingExecutor.ticker != nil {
		go func() {
			for _ = range bufferPaddingExecutor.ticker.C {
				bufferPaddingExecutor.paddingBuffer()
			}
		}()
	}

}

func (bufferPaddingExecutor BufferPaddingExecutor) isRunning() bool {

	return atomic.LoadInt64(&bufferPaddingExecutor.running) == 0
}

func (bufferPaddingExecutor BufferPaddingExecutor) asyncPadding() {
	t := NewTask(func() error {
		bufferPaddingExecutor.paddingBuffer()
		return nil
	})

	//开一个协程 不断的向 Pool 输送打印一条时间的task任务
	go func() {
		for {
			bufferPaddingExecutor.taskPool.EntryChannel <- t
		}
	}()

	//启动协程池p
	bufferPaddingExecutor.taskPool.Run()
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
			isFullRingBuffer = C.put(&bufferPaddingExecutor.ringbuffer, C.long(uid)) == 0
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

	if bufferPaddingExecutor.ticker != nil {
		bufferPaddingExecutor.ticker.Stop()
	}
}
