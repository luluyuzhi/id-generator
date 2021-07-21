package uidGenerator

/*
#cgo CFLAGS: -I${SRCDIR}/core/UidGenerator/include
#cgo LDFLAGS: -L${SRCDIR}/lib -luidgenerator
#include <RingBuffer.h>
*/
import "C"
import (
	"fmt"
	"unsafe"
)

const DEFAULT_BOOST_POWER = 3

type CachedUidGenerator struct {
	DefaultUidGenerator
	/** Spring properties */
	boostPower       int32
	paddingFactor    int32
	scheduleInterval int32

	// rejectedPutBufferHandler  RejectedPutBufferHandler
	// rejectedTakeBufferHandler RejectedTakeBufferHandler

	ringBuffer            C.struct_RingBuffer
	bufferPaddingExecutor BufferPaddingExecutor
}

func (cachedUidGenerator CachedUidGenerator) afterPropertiesSet() {
	(*DefaultUidGenerator)(unsafe.Pointer(&cachedUidGenerator)).afterPropertiesSet()

	// initialize RingBuffer & RingBufferPaddingExecutor
	cachedUidGenerator.initRingBuffer()
	fmt.Printf("Initialized RingBuffer successfully.")
}

func (cachedUidGenerator CachedUidGenerator) initRingBuffer() {
	var bufferSize = (cachedUidGenerator.bitsAllocator.getMaxSequence() + 1) << cachedUidGenerator.boostPower
	C.RingBufferInit(&cachedUidGenerator.ringBuffer, C.int32_t(bufferSize), C.int32_t(cachedUidGenerator.paddingFactor))
	fmt.Print("Initialized ring buffer size:{}, paddingFactor:{}", bufferSize, cachedUidGenerator.paddingFactor)
}

func (cachedUidGenerator CachedUidGenerator) getUID() int64 {

	return int64(C.take(&cachedUidGenerator.ringBuffer))
}
func (cachedUidGenerator CachedUidGenerator) parseUID(uid int64) string {
	return (*DefaultUidGenerator)(unsafe.Pointer(&cachedUidGenerator)).parseUID(uid)
}

func (cachedUidGenerator CachedUidGenerator) destroy() {
	// cachedUidGenerator.bufferPaddingExecutor.shutdown()
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
func (cachedUidGenerator CachedUidGenerator) setBoostPower(boostPower int32) {
	// Assert.isTrue(boostPower > 0, "Boost power must be positive!")
	cachedUidGenerator.boostPower = boostPower
}

// func (cachedUidGenerator CachedUidGenerator) setRejectedPutBufferHandler(rejectedPutBufferHandler RejectedPutBufferHandler) {
// 	// Assert.notNull(rejectedPutBufferHandler, "RejectedPutBufferHandler can't be null!")
// 	cachedUidGenerator.rejectedPutBufferHandler = rejectedPutBufferHandler
// }

// func (cachedUidGenerator CachedUidGenerator) setRejectedTakeBufferHandler(rejectedTakeBufferHandler RejectedTakeBufferHandler) {
// 	// Assert.notNull(rejectedTakeBufferHandler, "RejectedTakeBufferHandler can't be null!")
// 	cachedUidGenerator.rejectedTakeBufferHandler = rejectedTakeBufferHandler
// }

func (cachedUidGenerator CachedUidGenerator) setScheduleInterval(scheduleInterval int32) {
	// Assert.isTrue(scheduleInterval > 0, "Schedule interval must positive!")
	cachedUidGenerator.scheduleInterval = scheduleInterval
}
