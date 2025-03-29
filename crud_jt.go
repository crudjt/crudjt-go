package main

/*
#cgo linux,amd64 LDFLAGS: -L./native/linux/x86_64 -l store_jt -Wl,-rpath,.
#cgo linux,arm64 LDFLAGS: -L./native/linux/arm64 -l store_jt -Wl,-rpath,.

#cgo darwin,amd64 LDFLAGS: -L ./native/macos/x86_64 -l store_jt -Wl,-rpath,.
#cgo darwin,arm64 LDFLAGS: -L./native/macos/arm64 -l store_jt -Wl,-rpath,.

#cgo windows,amd64 LDFLAGS: -L./native/windows/x86_64 -l store_jt -Wl,-rpath,.
#cgo windows,arm64 LDFLAGS: -L./native/windows/arm64 -l store_jt -Wl,-rpath,.
#include "store_jt.h"
*/
import "C"
import (
	"github.com/vmihailenco/msgpack/v5"
	"unsafe"
  "encoding/json"
  // "fmt"
)

// Q аналог Ruby `q`
func Create(hash *map[string]interface{}, asdf, qwerty *int) string {
	// var asdfVal, qwertyVal int

  ttl_for_call := C.int64_t(-1)
  silence_read_for_call := C.int32_t(-1)

	// Якщо asdf == nil, створюємо змінну та передаємо її адресу
  if asdf != nil {
    asdf64 := int64(*asdf)

		ttl_for_call = C.int64_t(asdf64)
	}
	if qwerty != nil {
    qwerty32 := int32(*qwerty)

		silence_read_for_call = C.int32_t(qwerty32)
	}
  //
	// Сериализуем в MessagePack
	data, err := msgpack.Marshal(*hash)
	if err != nil {
		return ""
	}

	// Викликаємо C-функцію
	ptr := C.__create(
		(*C.uchar)(unsafe.Pointer(&data[0])), C.size_t(len(data)),
		ttl_for_call, silence_read_for_call,
	)

	// Конвертуємо C-строку в Go-строку
	defer C.free(unsafe.Pointer(ptr))
	return C.GoString(ptr)
}

// W аналог Ruby `w`
func Read(value string) (map[string]interface{}, error) {
	cValue := C.CString(value)
	defer C.free(unsafe.Pointer(cValue))

	ptr := C.__read(cValue)
	if ptr == nil {
		return nil, nil
	}

  // // Виводимо результат перед конвертацією
	// rawStr := C.GoString(ptr)
	// fmt.Println("Raw C Output:", rawStr) // Вивід на екран

	defer C.free(unsafe.Pointer(ptr))

  // response := C.GoString(ptr)
	// if response == "" {
	// 	return nil, nil
	// }

  // response := C.GoString(ptr)
  // fmt.Println("Response:", response == "")

  // if C.GoString(ptr) == "" {
  //   return nil, nil
  // }

	// Декодируем JSON
	var result map[string]interface{}
	err := json.Unmarshal([]byte(C.GoString(ptr)), &result)

  if len(result) == 0 {
		return nil, nil
	}

	return nil, err
}

// E аналог Ruby `e`
func Update(value string, hash *map[string]interface{}, asdf, qwerty *int) bool {
  ttl_for_call := C.int64_t(-1)
  silence_read_for_call := C.int32_t(-1)

  // Якщо asdf == nil, створюємо змінну та передаємо її адресу
  if asdf != nil {
    asdf64 := int64(*asdf)

    ttl_for_call = C.int64_t(asdf64)
  }
  if qwerty != nil {
    qwerty32 := int32(*qwerty)

    silence_read_for_call = C.int32_t(qwerty32)
  }

  // Сериализуем в MessagePack
  data, err := msgpack.Marshal(*hash)
  if err != nil {
    return false
  }

	cValue := C.CString(value)
	defer C.free(unsafe.Pointer(cValue))

	ptr := C.__update(cValue, (*C.uchar)(unsafe.Pointer(&data[0])), C.size_t(len(data)), ttl_for_call, silence_read_for_call)

  // defer C.free(unsafe.Pointer(ptr))

	return ptr == 1
}

// R аналог Ruby `r`
func Delete(value string) bool {
	cValue := C.CString(value)
	defer C.free(unsafe.Pointer(cValue))

	ptr := C.__delete(cValue)
	return ptr == 1
}
