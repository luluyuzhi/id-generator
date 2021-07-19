package uidGenerator

import (
	"fmt"
	"time"
)

type UidGeneratorImpler interface {
	UidGenerator
	afterPropertiesSet()
}

type WorkerIdAssigner interface {
	assignWorkerId() int64
}

const EPOCHSTR = "2006-01-02"
const TIMEBITS = 28
const WORKERBITS = 22
const SEQBITS = 13

type DefaultUidGenerator struct {
	/** Bits allocate */
	timeBits   int32
	workerBits int32
	seqBits    int32
	/** Customer epoch, unit as second. For example 2016-05-20 (ms: 1463673600000)*/
	epochStr      string
	epochSeconds  int64
	bitsAllocator BitsAllocator
	workerId      int64

	sequence   int64
	lastSecond int64

	workerIdAssigner WorkerIdAssigner
}

func NewDefaultUidGenerator() DefaultUidGenerator {
	var uidGenerator DefaultUidGenerator
	uidGenerator.timeBits = TIMEBITS
	uidGenerator.workerBits = WORKERBITS
	uidGenerator.seqBits = SEQBITS

	uidGenerator.epochSeconds = 1463673600000
	uidGenerator.epochStr = EPOCHSTR

	uidGenerator.sequence = 0
	uidGenerator.lastSecond = -1
	return uidGenerator
}

func (uidGenerator DefaultUidGenerator) afterPropertiesSet() {
	var bitsAllocator = NewBitsAllocator(uidGenerator.timeBits, uidGenerator.workerBits, uidGenerator.seqBits)
	uidGenerator.workerId = uidGenerator.workerIdAssigner.assignWorkerId()
	if uidGenerator.workerId > bitsAllocator.getMaxWorkerId() {
		fmt.Printf("Worker id {}  exceeds the max {}", uidGenerator.workerId, bitsAllocator.getMaxWorkerId())
	}

	fmt.Printf("Initialized bits(1, {}, {}, {}) for workerID:{}", uidGenerator.timeBits,
		uidGenerator.workerBits,
		uidGenerator.seqBits,
		uidGenerator.workerId)
}

func (uidGenerator DefaultUidGenerator) getUID() int64 {
	// try {
	// 	return uidGenerator.nextId();
	// } catch (Exception e) {
	// 	LOGGER.error("Generate unique id exception. ", e);
	// 	throw new UidGenerateException(e);
	// }
	return 1
}

func (uidGenerator DefaultUidGenerator) parseUID(uid int64) string {
	var totalBits int32 = TOTAL_BITS
	var signBits int32 = uidGenerator.bitsAllocator.getSignBits()
	var timestampBits int32 = uidGenerator.bitsAllocator.getTimestampBits()
	var workerIdBits int32 = uidGenerator.bitsAllocator.getWorkerIdBits()
	var sequenceBits int32 = uidGenerator.bitsAllocator.getSequenceBits()

	// parse UID
	var sequence int64 = (uid << (totalBits - sequenceBits)) >> (totalBits - sequenceBits)
	var workerId int64 = (uid << (timestampBits + signBits)) >> (totalBits - workerIdBits)
	var deltaSeconds int64 = uid >> (workerIdBits + sequenceBits)

	timeObj := time.Unix(uidGenerator.epochSeconds+deltaSeconds, 0)
	thatTimeStr := timeObj.String()

	// format as string
	return fmt.Sprintln("{\"UID\":\"%d\",\"timestamp\":\"%s\",\"workerId\":\"%d\",\"sequence\":\"%d\"}",
		uid, thatTimeStr, workerId, sequence)
}

func (uidGenerator DefaultUidGenerator) nextId() int64 {
	var currentSecond int64 = uidGenerator.getCurrentSecond()

	// Clock moved backwards, refuse to generate uid
	if currentSecond < uidGenerator.lastSecond {
		refusedSeconds := uidGenerator.lastSecond - currentSecond
		fmt.Printf("Clock moved backwards. Refusing for %d seconds", refusedSeconds)
	}

	// At the same second, increase sequence
	if currentSecond == uidGenerator.lastSecond {
		uidGenerator.sequence = (uidGenerator.sequence + 1) & uidGenerator.bitsAllocator.getMaxSequence()
		// Exceed the max sequence, we wait the next second to generate uid
		if uidGenerator.sequence == 0 {
			currentSecond = uidGenerator.getNextSecond(uidGenerator.lastSecond)
		}

		// At the different second, sequence restart from zero
	} else {
		uidGenerator.sequence = 0
	}

	uidGenerator.lastSecond = currentSecond

	// Allocate bits for UID
	return uidGenerator.bitsAllocator.allocate(currentSecond-uidGenerator.epochSeconds, uidGenerator.workerId, uidGenerator.sequence)
}

func (uidGenerator DefaultUidGenerator) getNextSecond(lastTimestamp int64) int64 {
	var timestamp = uidGenerator.getCurrentSecond()
	for {
		if timestamp <= lastTimestamp {
			timestamp = uidGenerator.getCurrentSecond()
			break
		}
	}

	return timestamp
}

func (uidGenerator DefaultUidGenerator) getCurrentSecond() int64 {
	currentSecond := time.Now().Unix()
	if currentSecond-uidGenerator.epochSeconds > uidGenerator.bitsAllocator.getMaxDeltaSeconds() {
		fmt.Printf("Timestamp bits is exhausted. Refusing UID generate. Now: {}", currentSecond)
	}

	return currentSecond
}

func (uidGenerator DefaultUidGenerator) setWorkerIdAssigner(workerIdAssigner WorkerIdAssigner) {
	uidGenerator.workerIdAssigner = workerIdAssigner
}
func (uidGenerator DefaultUidGenerator) setTimeBits(timeBits int32) {
	if timeBits > 0 {
		uidGenerator.timeBits = timeBits
	}
}

// func (uidGenerator DefaultUidGenerator) setEpochStr(epochStr string) {
// 	if StringUtils.isNotBlank(epochStr) {
// 		uidGenerator.epochStr = epochStr
// 		uidGenerator.epochSeconds = TimeUnit.MILLISECONDS.toSeconds(DateUtils.parseByDayPattern(epochStr).getTime())
// 	}
// }
