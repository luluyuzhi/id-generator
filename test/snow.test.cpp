
#include "snow.hpp"
#include <iostream>
#include <bitset>
#include <thread>
#include <vector>

int main()
{

    SnowflakeIdWorker snowflakeIdWorker;
    snowflakeIdWorkerInit(&snowflakeIdWorker, 1, 1);

    std::vector<std::thread> threads;
    pthread_mutex_t mutex;
    pthread_mutex_init(&mutex, nullptr);
    for (int thread = 0; thread < 10; thread++)
    {
        threads.push_back(std::thread([&snowflakeIdWorker, &mutex]()
                                      {
                                          for (long count = 0; count < 20; count++)
                                          {
                                              std::cout << std::bitset<64>(nextId(&snowflakeIdWorker, &mutex)) << std::endl;
                                          }
                                      }));
    }

    for (auto &t : threads)
    {
        t.join();
    }

    return 0;
}