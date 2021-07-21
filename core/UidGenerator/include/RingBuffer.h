#ifndef UIDGENERATOR_RINGBUFFER_
#define UIDGENERATOR_RINGBUFFER_

#include <inttypes.h>
#include <stdatomic.h>
#include <stdlib.h>

struct RingBuffer
{
    int64_t bufferSize;
    int64_t indexMask;
    atomic_long tail;
    atomic_long cursor;
    int64_t *slots;
    atomic_long *flags;
    int32_t paddingThreshold;
};

/**
 * Constructor with buffer size & padding factor
 * 
 * @param bufferSize must be positive & a power of 2
 * @param paddingFactor percent in (0 - 100). When the count of rest available UIDs reach the threshold, it will trigger padding buffer<br>
 *        Sample: paddingFactor=20, bufferSize=1000 -> threshold=1000 * 20 /100,  
 *        padding buffer will be triggered when tail-cursor<threshold
 */
void RingBufferInit(struct RingBuffer *ringBuffer, int32_t bufferSize, int32_t paddingFactor);

void RingBufferInitWithoutFactory(struct RingBuffer *ringBuffer, int32_t bufferSize);

/**
 * Put an UID in the ring & tail moved<br>
 * We use 'synchronized' to guarantee the UID fill in slot & publish new tail sequence as atomic operations<br>
 * 
 * <b>Note that: </b> It is recommended to put UID in a serialize way, cause we once batch generate a series UIDs and put
 * the one by one into the buffer, so it is unnecessary put in multi-threads
 *
 * @param uid
 * @return false means that the buffer is full, apply {@link RejectedPutBufferHandler}
 */
int put(struct RingBuffer *ringBuffer, long uid);

/**
 * Take an UID of the ring at the next cursor, this is a lock free operation by using atomic cursor<p>
 * 
 * Before getting the UID, we also check whether reach the padding threshold, 
 * the padding buffer operation will be triggered in another thread<br>
 * If there is no more available UID to be taken, the specified {@link RejectedTakeBufferHandler} will be applied<br>
 * 
 * @return UID
 * @throws IllegalStateException if the cursor moved back
 */
long take(struct RingBuffer *ringBuffer);

long getTail(struct RingBuffer *ringBuffer);

long getCursor(struct RingBuffer *ringBuffer);

#endif // UIDGENERATOR_RINGBUFFER_