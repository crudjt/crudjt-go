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

#ifdef __cplusplus
}
#endif

#endif // STORE_JT_H
