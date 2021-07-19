package uidGenerator

const TOTAL_BITS = 1 << 6

type BitsAllocator struct {
	/**
	 * Bits for [sign-> second-> workId-> sequence]
	 */
	signBits      int32 // 1
	timestampBits int32 // custom
	workerIdBits  int32 //
	sequenceBits  int32 //

	/**
	 * Max value for workId & sequence
	 */
	maxDeltaSecond int64 // maxDeltaSecond
	maxWorkerId    int64
	maxSequence    int64 // maxSequ
	/**
	 * Shift for timestamp & workerId
	 */
	timestampShift int32
	workerIdShift  int32
}

func NewBitsAllocator(timestampBits int32, workerIdBits int32, sequenceBits int32) BitsAllocator {

	var bitsAllocator BitsAllocator
	bitsAllocator.signBits = 1
	// var allocateTotalBits int32 = bitsAllocator.signBits + timestampBits + workerIdBits + sequenceBits
	bitsAllocator.timestampBits = timestampBits
	bitsAllocator.workerIdBits = workerIdBits
	bitsAllocator.sequenceBits = sequenceBits

	bitsAllocator.maxDeltaSecond = ^(-1 << timestampBits)
	bitsAllocator.maxWorkerId = ^(-1 << workerIdBits)
	bitsAllocator.maxSequence = ^(-1 << sequenceBits)

	// initialize shift
	bitsAllocator.timestampShift = workerIdBits + sequenceBits
	bitsAllocator.workerIdShift = sequenceBits
	return bitsAllocator
}

func (bitsAllocator BitsAllocator) allocate(deltaSeconds int64, workerId int64, sequence int64) int64 {
	return (deltaSeconds << bitsAllocator.timestampShift) | (workerId << bitsAllocator.workerIdShift) | sequence
}

func (bitsAllocator BitsAllocator) getTimestampBits() int32 {

	return bitsAllocator.timestampBits
}

func (bitsAllocator BitsAllocator) getSignBits() int32 {

	return bitsAllocator.signBits
}

func (bitsAllocator BitsAllocator) getWorkerIdBits() int32 {
	return bitsAllocator.workerIdBits
}

func (bitsAllocator BitsAllocator) getMaxWorkerId() int64 {

	return bitsAllocator.maxWorkerId
}

func (bitsAllocator BitsAllocator) getSequenceBits() int32 {

	return bitsAllocator.sequenceBits
}

func (bitsAllocator BitsAllocator) getMaxDeltaSeconds() int64 {

	return bitsAllocator.maxDeltaSecond
}

func (bitsAllocator BitsAllocator) getMaxSequence() int64 {

	return bitsAllocator.maxSequence
}
