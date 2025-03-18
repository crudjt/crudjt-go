#ifndef RUSTLIB_H
#define RUSTLIB_H

#include <stdint.h> // Для роботи з цілими числами

// Оголошення функцій, експортованих з Rust через extern "C"
#ifdef __cplusplus
extern "C" {
#endif

// Приклад функції, яка приймає ціле число і повертає ціле число
int32_t my_function(int32_t value);

// Якщо вам потрібно передати вказівники, то використовуйте uint8_t* (байти) або інші типи для роботи з пам'яттю
// Приклад функції, яка приймає рядок (char*) і повертає інший рядок
char* process_string(const char* input);

// Якщо у вас є функція з булевим типом, то можете використовувати int для значень true (1) і false (0)
int is_valid(int32_t value);

#ifdef __cplusplus
}
#endif

#endif // RUSTLIB_H
