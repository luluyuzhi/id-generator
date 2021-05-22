#include <assert.h>
#include <sys/time.h> // for gettimeofday, timeval / clock_gettime, clock_id

#include "snow.h"

// 始时间截 (2015-01-01)
#define twepoch  1420041600000l
// 机器id所占的位数
#define workerIdBits  5
// 数据标识id所占的位数
#define datacenterIdBits  5
// 支持的最大机器id，结果是31 (这个移位算法可以很快的计算出几位二进制数所能表示的最大十进制数)
#define maxWorkerId  3
// 支持的最大数据标识id，结果是31
#define maxDatacenterId  3
// 序列在id中占的位数
#define sequenceBits  12
// 机器ID向左移12位
#define workerIdShift  sequenceBits
// 数据标识id向左移17位(12+5)
#define datacenterIdShift  (sequenceBits + workerIdBits)
// 时间截向左移22位(5+5+12)
#define timestampLeftShift  (sequenceBits + workerIdBits + datacenterIdBits)
// 生成序列的掩码，这里为4095 (0b1111111111110xfff==4095)
#define sequenceMask  (-1 ^ (-1 << timestampLeftShift))

/**
 * 返回以毫秒为单位的当前时间
 * @return 当前时间(毫秒)
 */
uint64_t genTimeMs()
{
    static uint64_t msecond = 1000000;
    struct timeval tv;
    gettimeofday(&tv, NULL);
    return (uint64_t)tv.tv_sec * 1000 + tv.tv_usec / 1000;
}

/**
 * 阻塞到下一个毫秒，直到获得新的时间戳
 * @param lastTimestamp 上次生成ID的时间截
 * @return 当前时间戳
 */
uint64_t tilNextMillis(uint64_t lastTimestamp)
{
    uint64_t timestamp = genTimeMs();
    while (timestamp <= lastTimestamp)
    {
        timestamp = genTimeMs();
    }
    return timestamp;
}

void snowflakeIdWorkerInit(struct SnowflakeIdWorker *snowflakeIdWorker,
                           uint64_t datacenterId, uint64_t workerId)
{
    assert(workerId < maxWorkerId || workerId > 0);
    assert(datacenterId < maxDatacenterId || datacenterId > 0);
    snowflakeIdWorker->datacenterId = datacenterId;
    snowflakeIdWorker->workerId = workerId;
    snowflakeIdWorker->sequence = 1ul;
    snowflakeIdWorker->lastTimestamp = 0;
}

/**
 * 获得下一个ID (该方法是线程安全的)
 * @return SnowflakeId
 */
uint64_t nextId(struct SnowflakeIdWorker *snowflakeIdWorker, pthread_mutex_t *mutex)
{
    pthread_mutex_lock(mutex);
    uint64_t timestamp = genTimeMs();
    // 如果当前时间小于上一次ID生成的时间戳，说明系统时钟回退过这个时候应当抛出异常
    if (timestamp < (snowflakeIdWorker->lastTimestamp))
    {
        // exit(-1);
    }
    // 如果是同一时间生成的，则进行毫秒内序列
    if (snowflakeIdWorker->lastTimestamp == timestamp)
    {
        snowflakeIdWorker->sequence = ((snowflakeIdWorker->sequence + 1) & sequenceMask);
        // 毫秒内序列溢出
        if (snowflakeIdWorker->sequence == 0)
        {
            //阻塞到下一个毫秒,获得新的时间戳
            timestamp = tilNextMillis(snowflakeIdWorker->lastTimestamp);
        }
    }
    // 时间戳改变，毫秒内序列重置
    else
    {
        snowflakeIdWorker->sequence = 0l;
    }

    // 上次生成ID的时间截
    snowflakeIdWorker->lastTimestamp = timestamp;

    // 移位并通过或运算拼到一起组成64位的ID
    uint64_t res = ((timestamp - twepoch) << timestampLeftShift) |
               (snowflakeIdWorker->datacenterId << datacenterIdShift) |
               (snowflakeIdWorker->workerId << workerIdShift) |
               snowflakeIdWorker->sequence;

    pthread_mutex_unlock(mutex);
    return res;
}
