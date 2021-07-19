int main(int argc, char const *argv[])
{
    struct RingBuffer ringBuffer;
    RingBufferInit(&ringBuffer, 100, 10);
    RingBufferInitWithoutFactory(&ringBuffer, 1000);
    printf("%ld", put(&ringBuffer, 1));
    return 0;
}
