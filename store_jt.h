#ifndef STORE_JT_H
#define STORE_JT_H

#include <stdint.h>
#include <stddef.h>
#include <stdbool.h>
#include <stdlib.h>
#include <stdio.h>

#ifdef __cplusplus
extern "C" {
#endif

const char* start_store_jt(const char *key, const char *path_to_db);

const char* __create(const uint8_t* data, size_t len, int64_t ttl, int32_t silence_read);

const char* __read(const char* token);

int __update(const char* token, const uint8_t* data, size_t len, int64_t ttl, int32_t silence_read);

int __delete(const char* token);

#ifdef __cplusplus
}
#endif

#endif // STORE_JT_H
