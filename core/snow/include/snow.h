// base on https://zhuanlan.zhihu.com/p/85837641
#include <inttypes.h>
#include <pthread.h>

struct SnowflakeIdWorker
{
    // 数据中心ID(0~31)
    uint64_t datacenterId;
    // 工作机器ID(0~31)
    uint64_t workerId;
    // 毫秒内序列(0~4095)
    uint64_t sequence;
    // 上次生成ID的时间截
    uint64_t lastTimestamp;
};

void snowflakeIdWorkerInit(struct SnowflakeIdWorker *snowflakeIdWorker,
                           uint64_t datacenterId, uint64_t workerId);

/**
 * 获得下一个ID (该方法是线程安全的)
 * @return SnowflakeId
 */
uint64_t nextId(struct SnowflakeIdWorker *snowflakeIdWorker, pthread_mutex_t *mutex);
