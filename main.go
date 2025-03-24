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
	"fmt"
	"unsafe"
)

func main() {
	str := "fsdfsfsdfsdfsdfsdf"

	// Конвертація Go-рядка в C-string
	cStr := C.CString(str)
	defer C.free(unsafe.Pointer(cStr)) // Звільнення пам'яті після виклику

	// Виклик функції з Rust
	C.encrypted_key(cStr)

	fmt.Println("Called Rust function yo successfully!")
}

// func main() {
//     value := C.my_function(10)
//     fmt.Println("Result:", value)
// }
