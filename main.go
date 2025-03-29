package main

import (
	"C"
	"fmt"
)

func main() {
	// Створюємо хеш (map)
	data := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
		"key3": true,
	}

	// Викликаємо функцію Q
	ttl := 100
	silence_read := 17
	result := Create(&data, &ttl, &silence_read)

	// Виводимо результат
	fmt.Println("Result from Create:", result)
	fmt.Println(Read(result))

	updated_data := map[string]interface{}{
		"user_id": 42,
		"role": 11,
	}
	new_ttl := 500
	Update(result, &updated_data, &new_ttl, nil)
	fmt.Println("Result from Update:")
	fmt.Println(Read(result))

	Delete(result)
	fmt.Println("Result for Delete:")
	fmt.Println(Read(result))

	fmt.Println("Called Rust function yo successfully!")
}

// func main() {
//     value := C.my_function(10)
//     fmt.Println("Result:", value)
// }
