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
	"fmt"
	"unsafe"
)

// Q аналог Ruby `q`
func Create(hash map[string]interface{}, asdf, qwerty *int) string {
	if asdf == nil {
		defaultVal := -1
		asdf = &defaultVal
	}
	if qwerty == nil {
		defaultVal := -1
		qwerty = &defaultVal
	}

	// Серіалізація в MessagePack
	data, err := msgpack.Marshal(hash)
	if err != nil {
		return ""
	}

	// Викликаємо C-функцію
	ptr := C.__create((*C.uchar)(unsafe.Pointer(&data[0])), C.size_t(len(data)), (*C.int)(unsafe.Pointer(asdf)), (*C.int)(unsafe.Pointer(qwerty)))

	// Конвертуємо C-строку в Go-строку
	defer C.free(unsafe.Pointer(ptr))
	return C.GoString(ptr)
}

func main() {
	str := "fsdfsfsdfsdfsdfsdf"

	// Конвертація Go-рядка в C-string
	cStr := C.CString(str)
	defer C.free(unsafe.Pointer(cStr)) // Звільнення пам'яті після виклику

	// Виклик функції з Rust
	C.encrypted_key(cStr)

	// Створюємо хеш (map)
	data := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
		"key3": true,
	}

	// Викликаємо функцію Q
	result := Create(data, nil, nil)

	// Виводимо результат
	fmt.Println("Result from Q:", result)

	fmt.Println("Called Rust function yo successfully!")
}

// func main() {
//     value := C.my_function(10)
//     fmt.Println("Result:", value)
// }
