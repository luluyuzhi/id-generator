package uidGenerator

/*
#cgo CFLAGS: -I../core/UidGenerator/include
#cgo LDFLAGS: -L../lib -luidgenerator
#include <RingBuffer.h>
*/
import "C"
import (
	"errors"
	"fmt"
	"unsafe"
)

const (
	DEFAULT_BOOST_POWER     = 3
	START_POINT             = -1
	CAN_PUT_FLAG            = 0
	CAN_TAKE_FLAG           = 1
	DEFAULT_PADDING_PERCENT = 50
)

type CachedUidGenerator struct {
	DefaultUidGenerator

	boostPower       int32
	paddingFactor    int32
	scheduleInterval int64 // set to 0 if not using schedule

	ringBuffer            C.struct_RingBuffer
	bufferPaddingExecutor *BufferPaddingExecutor
}

func New(workerIdAssigner WorkerIdAssigner) *CachedUidGenerator {
	var cachedUidGenerator CachedUidGenerator

	(*DefaultUidGenerator)(unsafe.Pointer(&cachedUidGenerator)).init()
	cachedUidGenerator.boostPower = DEFAULT_BOOST_POWER
	cachedUidGenerator.paddingFactor = DEFAULT_PADDING_PERCENT

	cachedUidGenerator.SetWorkerIdAssigner(workerIdAssigner)

	(*DefaultUidGenerator)(unsafe.Pointer(&cachedUidGenerator)).afterPropertiesSet()

	// initialize RingBuffer & RingBufferPaddingExecutor
	cachedUidGenerator.initRingBuffer()
	fmt.Printf("Initialized RingBuffer successfully.")
	return &cachedUidGenerator
}

func (cachedUidGenerator *CachedUidGenerator) initRingBuffer() {
	var bufferSize = (cachedUidGenerator.bitsAllocator.getMaxSequence() + 1) << cachedUidGenerator.boostPower
	C.RingBufferInit(&cachedUidGenerator.ringBuffer, C.int32_t(bufferSize), C.int32_t(cachedUidGenerator.paddingFactor))
	fmt.Printf("Initialized ring buffer size:%d, paddingFactor:%d\n", bufferSize, cachedUidGenerator.paddingFactor)
	var usingSchedule = (cachedUidGenerator.scheduleInterval != 0)
	cachedUidGenerator.bufferPaddingExecutor = NewBufferPaddingExecutor(&cachedUidGenerator.ringBuffer, true)

	cachedUidGenerator.bufferPaddingExecutor.bufferedUidProvider = FuncCaller(cachedUidGenerator.nextIdsForOneSecond)
	if usingSchedule {
		cachedUidGenerator.bufferPaddingExecutor.setScheduleInterval(cachedUidGenerator.scheduleInterval)
	}

	fmt.Printf("Initialized BufferPaddingExecutor. Using schdule:%t, interval:%+d\n", usingSchedule, cachedUidGenerator.scheduleInterval)

	// fill in all slots of the RingBuffer
	cachedUidGenerator.bufferPaddingExecutor.paddingBuffer()

	// start buffer padding threads
	cachedUidGenerator.bufferPaddingExecutor.start()
}

func (cachedUidGenerator *CachedUidGenerator) GetUID() (int64, error) {

	takeResult := int64(C.take(&cachedUidGenerator.ringBuffer))
	if takeResult == 0 {
		cachedUidGenerator.bufferPaddingExecutor.asyncPadding()
	}

	if takeResult == -1 {
		return -1, errors.New("unsolve error generator")
	}
	return takeResult, nil
}

func (cachedUidGenerator CachedUidGenerator) ParseUID(uid int64) string {
	return (*DefaultUidGenerator)(unsafe.Pointer(&cachedUidGenerator)).ParseUID(uid)
}

func (cachedUidGenerator CachedUidGenerator) nextIdsForOneSecond(currentSecond int64) []int64 {
	// Initialize result list size of (max sequence + 1)
	var listSize = cachedUidGenerator.bitsAllocator.getMaxSequence() + 1
	var uidList []int64

	// Allocate the first sequence of the second, the others can be calculated with the offset
	firstSeqUid := cachedUidGenerator.bitsAllocator.allocate(currentSecond-cachedUidGenerator.epochSeconds, cachedUidGenerator.workerId, 0)
	var offset int64 = 0
	for ; offset < listSize; offset++ {
		uidList = append(uidList, firstSeqUid+offset)
	}
	return uidList
}

/**
 * Setters for spring property
 */
func (cachedUidGenerator *CachedUidGenerator) SetBoostPower(boostPower int32) {
	// assert.Equal(boostPower > 0, true, "Boost power must be positive!")
	cachedUidGenerator.boostPower = boostPower
}

func (cachedUidGenerator *CachedUidGenerator) SetScheduleInterval(scheduleInterval int64) {
	// assert.Equal(scheduleInterval > 0, true, "Schedule interval must positive!")
	cachedUidGenerator.scheduleInterval = scheduleInterval

}

func Destroy(cachedUidGenerator *CachedUidGenerator) {
	cachedUidGenerator.bufferPaddingExecutor.shutdown()
}

// runtime.SetFinalizer(r, destroy)
