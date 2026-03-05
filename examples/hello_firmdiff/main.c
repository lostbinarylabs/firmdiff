#include <stdio.h>
#include <stdint.h>

static uint32_t compute(uint32_t x)
{
    uint32_t s = 0;
    for (uint32_t i = 0; i < x; i++)
    {
        s += i * 3u;
    }
    return s;
}

int main(void)
{
    printf("hello firmdiff\n");
    printf("result=%u\n", compute(1000));
    return 0;
}