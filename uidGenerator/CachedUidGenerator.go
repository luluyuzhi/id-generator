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

const _DEFAULT_BOOST_POWER = 3
const START_POINT = -1
const CAN_PUT_FLAG = 0
const CAN_TAKE_FLAG = 1
const DEFAULT_PADDING_PERCENT = 50

type CachedUidGenerator struct {
	DefaultUidGenerator

	boostPower       int32
	paddingFactor    int32
	scheduleInterval int64

	// rejectedPutBufferHandler  RejectedPutBufferHandler
	// rejectedTakeBufferHandler RejectedTakeBufferHandler

	ringBuffer            C.struct_RingBuffer
	bufferPaddingExecutor *BufferPaddingExecutor
}

func New() *CachedUidGenerator {
	var cachedUidGenerator CachedUidGenerator

	cachedUidGenerator.boostPower = _DEFAULT_BOOST_POWER
	cachedUidGenerator.paddingFactor = DEFAULT_PADDING_PERCENT
	(*DefaultUidGenerator)(unsafe.Pointer(&cachedUidGenerator)).afterPropertiesSet()

	// initialize RingBuffer & RingBufferPaddingExecutor
	cachedUidGenerator.initRingBuffer()
	fmt.Printf("Initialized RingBuffer successfully.")
	return &cachedUidGenerator
}

func (cachedUidGenerator *CachedUidGenerator) initRingBuffer() {
	var bufferSize = (cachedUidGenerator.bitsAllocator.getMaxSequence() + 1) << cachedUidGenerator.boostPower
	C.RingBufferInit(&cachedUidGenerator.ringBuffer, C.int32_t(bufferSize), C.int32_t(cachedUidGenerator.paddingFactor))
	fmt.Print("Initialized ring buffer size:{}, paddingFactor:{}", bufferSize, cachedUidGenerator.paddingFactor)
	var usingSchedule = (cachedUidGenerator.scheduleInterval != 0)
	cachedUidGenerator.bufferPaddingExecutor = NewBufferPaddingExecutor(&cachedUidGenerator.ringBuffer, true)
	cachedUidGenerator.bufferPaddingExecutor.bufferedUidProvider = cachedUidGenerator.nextIdsForOneSecond
	if usingSchedule {
		cachedUidGenerator.bufferPaddingExecutor.setScheduleInterval(cachedUidGenerator.scheduleInterval)
	}
	fmt.Printf("Initialized BufferPaddingExecutor. Using schdule:{}, interval:{}", usingSchedule, cachedUidGenerator.scheduleInterval)
	// // set rejected put/take handle policy
	// this.ringBuffer.setBufferPaddingExecutor(bufferPaddingExecutor);
	// if (rejectedPutBufferHandler != null) {
	// 	this.ringBuffer.setRejectedPutHandler(rejectedPutBufferHandler);
	// }
	// if (rejectedTakeBufferHandler != null) {
	// 	this.ringBuffer.setRejectedTakeHandler(rejectedTakeBufferHandler);
	// }

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
	return (*DefaultUidGenerator)(unsafe.Pointer(&cachedUidGenerator)).parseUID(uid)
}

func (cachedUidGenerator CachedUidGenerator) destroy() {
	cachedUidGenerator.bufferPaddingExecutor.shutdown()
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
