#ifndef UIDGENERATOR_RINGBUFFER_
#define UIDGENERATOR_RINGBUFFER_

#include <inttypes.h>
#include <stdatomic.h>
#include <stdlib.h>

struct RingBuffer
{
    int32_t bufferSize;
    int32_t indexMask;
    int64_t *slots;
    atomic_long *flags;
};

void RingBufferInit(struct RingBuffer *ringBuffer, int32_t bufferSize)
{
    ringBuffer->bufferSize = bufferSize;
    ringBuffer->indexMask = bufferSize - 1;
    ringBuffer->slots = (int64_t *)malloc(sizeof(int64_t) * bufferSize);
    ringBuffer->flags = (atomic_long *)malloc(sizeof(atomic_long) * bufferSize);
}

static void flagsInit(atomic_long *flags, int32_t bufferSize)
{
    int32_t flagsIndex = 0;

    for (; flagsIndex < bufferSize; flagsIndex++)
    {
        atomic_init(&flags[flagsIndex], 0);
    }
}

#endif // UIDGENERATOR_RINGBUFFER_