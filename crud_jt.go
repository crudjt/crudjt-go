package crud_jt

/*
#cgo linux,amd64 LDFLAGS: -L${SRCDIR}/native/linux/x86_64 -l store_jt -Wl,-rpath,${SRCDIR}/native/linux/x86_64
#cgo linux,arm64 LDFLAGS: -L${SRCDIR}/native/linux/arm64 -l store_jt -Wl,-rpath,${SRCDIR}/native/linux/arm64

#cgo darwin,amd64 LDFLAGS: -L${SRCDIR}/native/macos/x86_64 -lstore_jt -Wl,-rpath,${SRCDIR}/native/macos/x86_64
#cgo darwin,arm64 LDFLAGS: -L${SRCDIR}/native/macos/arm64 -l store_jt -Wl,-rpath,${SRCDIR}/native/macos/arm64
#include "store_jt.h"
*/
import "C"
import (
	"github.com/vmihailenco/msgpack/v5"
	"unsafe"
  "encoding/json"
  "fmt"
)

// Config тримає налаштування модуля
type Config struct {
    EncryptedKey string
    StoreJtPath string
}

// приватна змінна з конфігом
var config Config

var CacheInstance *LRUCache

// SetConfig встановлює глобальну конфігурацію модуля
func SetConfig(cfg Config) {
    config = cfg

		if cfg.StoreJtPath != "" {
			cstr := C.CString(cfg.StoreJtPath)
			defer C.free(unsafe.Pointer(cstr))
			C.__store_jt_path(cstr)
		}

		cstr := C.CString(cfg.EncryptedKey)
		defer C.free(unsafe.Pointer(cstr))
		C.__encrypted_key(cstr)
}

func init() {
	CacheInstance = NewLRUCache(OriginalRead)
}

// Q аналог Ruby `q`
func Create(hash *map[string]interface{}, asdf, qwerty *int) string {
	err := ValidateInsertion(hash, asdf, qwerty)
	if err != nil {
		fmt.Println("Create error:", err)

		panic("Ups") // mock
		return ""
	}

  ttl_for_call := C.int64_t(-1)
  silence_read_for_call := C.int32_t(-1)

	ttl_for_cache := -1
	silence_read_for_cache := -1

	// Якщо asdf == nil, створюємо змінну та передаємо її адресу
  if asdf != nil {
    asdf64 := int64(*asdf)

		ttl_for_call = C.int64_t(asdf64)
		ttl_for_cache = int(*asdf) + 1 // TODO: move to CacheInstance
	}
	if qwerty != nil {
    qwerty32 := int32(*qwerty)

		silence_read_for_call = C.int32_t(qwerty32)
		silence_read_for_cache = int(*qwerty)
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

	token := C.GoString(ptr)
	CacheInstance.Insert(token, *hash, ttl_for_cache, silence_read_for_cache)

	// Конвертуємо C-строку в Go-строку
	defer C.free(unsafe.Pointer(ptr))
	return token
}

func OriginalRead(value string) {
	cValue := C.CString(value)
	defer C.free(unsafe.Pointer(cValue))

	ptr := C.__read(cValue)
	defer C.free(unsafe.Pointer(ptr))
}

// W аналог Ruby `w`
func Read(value string) (map[string]interface{}, error) {
	read_err := ValidateToken(value)
	if read_err != nil {
		fmt.Println("Read error:", read_err)

		panic("Ups") // mock
	}

	output := CacheInstance.Get(value)
	if output != nil {
		return output, nil
	}

	cValue := C.CString(value)
	defer C.free(unsafe.Pointer(cValue))

	ptr := C.__read(cValue)
	defer C.free(unsafe.Pointer(ptr))

	if ptr == nil {
		return nil, nil
	}

  // // Виводимо результат перед конвертацією
	// rawStr := C.GoString(ptr)
	// fmt.Println("Raw C Output:", rawStr) // Вивід на екран


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
	CacheInstance.ForceInsert(value, result)

	return result, err
}

// E аналог Ruby `e`
func Update(value string, hash *map[string]interface{}, asdf, qwerty *int) bool {
	err := ValidateInsertion(hash, asdf, qwerty)
	if err != nil {
		fmt.Println("Update error:", err)

		panic("Ups")
	}

  ttl_for_call := C.int64_t(-1)
  silence_read_for_call := C.int32_t(-1)

	ttl_for_cache := -1
	silence_read_for_cache := -1

  // Якщо asdf == nil, створюємо змінну та передаємо її адресу
  if asdf != nil {
    asdf64 := int64(*asdf)

    ttl_for_call = C.int64_t(asdf64)
		ttl_for_cache = int(*asdf) + 1 // TODO: move to CacheInstance
  }
  if qwerty != nil {
    qwerty32 := int32(*qwerty)

    silence_read_for_call = C.int32_t(qwerty32)
		silence_read_for_cache = int(*qwerty)
  }

  // Сериализуем в MessagePack
  data, err := msgpack.Marshal(*hash)
  if err != nil {
    return false
  }

	cValue := C.CString(value)
	defer C.free(unsafe.Pointer(cValue))

	ptr := C.__update(cValue, (*C.uchar)(unsafe.Pointer(&data[0])), C.size_t(len(data)), ttl_for_call, silence_read_for_call)

	bool := ptr == 1
	if bool {
		CacheInstance.Insert(value, *hash, ttl_for_cache, silence_read_for_cache)
	}

  // defer C.free(unsafe.Pointer(ptr))

	return bool
}

// R аналог Ruby `r`
func Delete(value string) bool {
	delete_err := ValidateToken(value)
	if delete_err != nil {
		fmt.Println("Delete error:", delete_err)

		panic("Ups")
	}

	cValue := C.CString(value)
	defer C.free(unsafe.Pointer(cValue))

	ptr := C.__delete(cValue)
	CacheInstance.Delete(value)
	return ptr == 1
}
