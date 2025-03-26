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
void encrypted_key(const char *key);

const char* __create(const uint8_t* data, size_t len, const int* ttl, const int* silence_read);

const char* __read(const char* token);

const int* __update(const char* token, const uint8_t* data, size_t len, const int* ttl, const int* silence_read);

const int* __delete(const char* token);

#ifdef __cplusplus
}
#endif

#endif // STORE_JT_H
