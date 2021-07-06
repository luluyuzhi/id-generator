// #include "RingBuffer.h"

#include <inttypes.h>
#include <stdatomic.h>
#include <stdlib.h>

// for test
#include <stdio.h>

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

static void flagsInit(atomic_long *flags, int32_t bufferSize);

/**
 * Constructor with buffer size & padding factor
 * 
 * @param bufferSize must be positive & a power of 2
 * @param paddingFactor percent in (0 - 100). When the count of rest available UIDs reach the threshold, it will trigger padding buffer<br>
 *        Sample: paddingFactor=20, bufferSize=1000 -> threshold=1000 * 20 /100,  
 *        padding buffer will be triggered when tail-cursor<threshold
 */
void RingBufferInit(struct RingBuffer *ringBuffer, int32_t bufferSize, int32_t paddingFactor)
{
    ringBuffer->bufferSize = bufferSize;
    ringBuffer->indexMask = bufferSize - 1;
    ringBuffer->slots = (int64_t *)malloc(sizeof(int64_t) * bufferSize);
    ringBuffer->flags = (atomic_long *)malloc(sizeof(atomic_long) * bufferSize);

    ringBuffer->paddingThreshold = bufferSize * paddingFactor / 100;
}

void RingBufferInitWithoutFactory(struct RingBuffer *ringBuffer, int32_t bufferSize)
{
    const static int32_t DEFAULT_PADDING_PERCENT = 50;
    RingBufferInit(ringBuffer, bufferSize, DEFAULT_PADDING_PERCENT);
}

static void flagsInit(atomic_long *flags, int32_t bufferSize)
{
    int32_t flagsIndex = 0;

    for (; flagsIndex < bufferSize; flagsIndex++)
    {
        atomic_init(&flags[flagsIndex], 0);
    }
}

static int calSlotIndex(struct RingBuffer *ringBuffer, long sequence)
{
    return (int)(sequence & ringBuffer->indexMask);
}

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
int put(struct RingBuffer *ringBuffer, long uid)
{
    long currentTail = atomic_load(&ringBuffer->tail);
    long currentCursor = atomic_load(&ringBuffer->cursor);

    const static long START_POINT = -1;
    // tail catches the cursor, means that you can't put any cause of RingBuffer is full
    long distance = currentTail - (currentCursor == START_POINT ? 0 : currentCursor);
    if (distance == ringBuffer->bufferSize - 1)
    {
        // rejectedPutHandler.rejectPutBuffer(this, uid);
        printf("Rejected putting buffer for uid: %ld", uid);
        return -1;
    }

    // 1. pre-check whether the flag is CAN_PUT_FLAG
    int nextTailIndex = calSlotIndex(ringBuffer, currentTail + 1);
    const static long CAN_PUT_FLAG = 0L;
    if (atomic_load(&ringBuffer->flags[nextTailIndex]) != CAN_PUT_FLAG)
    {
        // rejectedPutHandler.rejectPutBuffer(this, uid);
        printf("Rejected putting buffer for uid: %ld", uid);
        return -1;
    }

    // 2. put UID in the next slot
    // 3. update next slot' flag to CAN_TAKE_FLAG
    // 4. publish tail with sequence increase by one
    ringBuffer->slots[nextTailIndex] = uid;
    const static long CAN_TAKE_FLAG = 0L;
    atomic_store(&ringBuffer->flags[nextTailIndex], (CAN_TAKE_FLAG));
    atomic_fetch_add(&ringBuffer->tail, 1);

    // The atomicity of operations above, guarantees by 'synchronized'. In another word,
    // the take operation can't consume the UID we just put, until the tail is published(tail.incrementAndGet())
    return 0;
}

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
long take(struct RingBuffer *ringBuffer)
{
    // spin get next available cursor
    long currentCursor = atomic_load(&ringBuffer->cursor);
    // long nextCursor = cursor.updateAndGet(old->old == tail.get() ? old : old + 1);
    long oldCursor = atomic_load(&ringBuffer->cursor);

    int nextCursor;
    {
        int prev;
        atomic_long newPrev;
        do
        {

            prev = atomic_load(&ringBuffer->cursor);

            if (prev == atomic_load(&ringBuffer->tail))
            {
                nextCursor = prev;
            }
            else
            {
                nextCursor = prev + 1;
            }
            atomic_init(&newPrev, prev);
        } while (!atomic_compare_exchange_strong(&ringBuffer->cursor, &newPrev, nextCursor));
    }

    long currentTail = atomic_load(&ringBuffer->tail);

    if (currentTail - nextCursor < ringBuffer->paddingThreshold)
    {

        // bufferPaddingExecutor.asyncPadding();
    }

    if (nextCursor == currentCursor)
    {

        // rejectedTakeHandler.rejectTakeBuffer(this);
    }

    int nextCursorIndex = calSlotIndex(ringBuffer, nextCursor);
    // Assert.isTrue(flags[nextCursorIndex].get() == CAN_TAKE_FLAG, "Curosr not in can take status");

    // 2. get UID from next slot
    // 3. set next slot flag as CAN_PUT_FLAG.
    long uid = ringBuffer->slots[nextCursorIndex];
    const static long CAN_PUT_FLAG = 0;
    atomic_store(&ringBuffer->flags[nextCursorIndex], CAN_PUT_FLAG);

    // Note that: Step 2,3 can not swap. If we set flag before get value of slot, the producer may overwrite the
    // slot with a new UID, and this may cause the consumer take the UID twice after walk a round the ring
    return uid;
}

int main(int argc, char const *argv[])
{
    struct RingBuffer ringBuffer;
    RingBufferInit(&ringBuffer, 100, 10);
    RingBufferInitWithoutFactory(&ringBuffer, 1000);
    printf("%ld", put(&ringBuffer, 1));
    return 0;
}
