#ifndef STORE_JT_H
#define STORE_JT_H

#include <stdint.h>  // Для сумісності з типами Rust
#include <stddef.h>  // Для size_t
#include <stdbool.h> // Для bool
#include <stdlib.h>  // Для malloc, free
#include <stdio.h>   // Для printf (якщо потрібно)

#ifdef __cplusplus
extern "C" {
#endif

// Оголошення функції з Rust
void __encrypted_key(const char *key);

void __store_jt_path(const char *path_to_db);

const char* __create(const uint8_t* data, size_t len, int64_t ttl, int32_t silence_read);

const char* __read(const char* token);

int __update(const char* token, const uint8_t* data, size_t len, int64_t ttl, int32_t silence_read);

int __delete(const char* token);

#ifdef __cplusplus
}
#endif

#endif // STORE_JT_H
