package main

import (
	"encoding/json"
	"fmt"
	"runtime"
	"time"
	// "math/rand"
)

func main() {
	// CacheInstance = NewLRUCache(OriginalRead)


	fmt.Printf("OS: %s\n", runtime.GOOS)
	fmt.Printf("CPU: %s\n", runtime.GOARCH)

	// without metadata
	fmt.Println("Checking without metadata")
	data := map[string]interface{}{"user_id": 42, "role": 11}
	expectedData := map[string]interface{}{"data": data}

	edData := map[string]interface{}{"user_id": 42, "role": 8}
	expectedEdData := map[string]interface{}{"data": edData}

	value := Create(&data, nil, nil)

	result, _ := Read(value)
	j1, _ := json.Marshal(result)
	j2, _ := json.Marshal(expectedData)
	fmt.Println(string(j1) == string(j2))
	fmt.Println(Update(value, &edData, nil, nil) == true)
	result2, _ := Read(value)

	j3, _ := json.Marshal(result2)
	j4, _ := json.Marshal(expectedEdData)
	fmt.Println(string(j3) == string(j4))
	fmt.Println(Delete(value) == true)
	result3, _ := Read(value)
	fmt.Println(result3 == nil)

	// with ttl
	fmt.Println("with ttl")
	ttl := 5
	valueWithttl := Create(&data, &ttl, nil)

	expectedttl := ttl
	for i := 0; i < ttl; i++ {
		expectedJSON, _ := json.Marshal(map[string]interface{}{"metadata": map[string]int{"ttl": expectedttl}, "data": data})
		result, _ := Read(valueWithttl)
		jsonValue, _ := json.Marshal(result)
		fmt.Println(string(jsonValue) == string(expectedJSON))
		expectedttl--
		time.Sleep(1 * time.Second)
	}
	output, _ := Read(valueWithttl)
	fmt.Println(output == nil)

	// when expired ttl
	fmt.Println("when expired ttl")
	ttl = 1
	value = Create(&data, &ttl, nil)
	time.Sleep(time.Duration(ttl) * time.Second)
	expired_ttl_output, _ := Read(value)
	fmt.Println(expired_ttl_output == nil)

	fmt.Println(Update(value, &data, nil, nil) == false)
	fmt.Println(Delete(value) == false)

	// with silence_read
	fmt.Println("with silence_read")
	silence_read := 6
	valueWithsilence_read := Create(&data, nil, &silence_read)
	// fmt.Println(Read(valueWithsilence_read))

	expectedsilence_read := silence_read - 1
	for i := 0; i < silence_read; i++ {
		expectedJSON, _ := json.Marshal(map[string]interface{}{"metadata": map[string]int{"silence_read": expectedsilence_read}, "data": data})
		sr_output, _ := Read(valueWithsilence_read)
		jsonValue, _ := json.Marshal(sr_output)
		fmt.Println(string(jsonValue) == string(expectedJSON))
		expectedsilence_read--
	}
	sr_output, _ := Read(valueWithsilence_read)
	fmt.Println(sr_output == nil)

	// with ttl and silence_read
	fmt.Println("Checking ttl and silence_read")

	data = map[string]interface{}{"user_id": 42, "role": 11}
	ttl = 5
	silence_read = ttl

	valueWithTtlAndsilence_read := Create(&data, &ttl, &silence_read)

	expectedttl = ttl
	expectedsilence_read = silence_read - 1

	// Проходимо через цикл і порівнюємо JSON-результати
	for i := 0; i < silence_read; i++ {
		// Формуємо очікувані значення для порівняння
		expectedJSON, _ := json.Marshal(map[string]interface{}{"metadata": map[string]int{"ttl": expectedttl, "silence_read": expectedsilence_read}, "data": data})
		ttl_and_sr_response, _ := Read(valueWithTtlAndsilence_read)
		jsonValue, _ := json.Marshal(ttl_and_sr_response)

		// Порівнюємо JSON
		fmt.Println(string(jsonValue) == string(expectedJSON))

		expectedttl--
		expectedsilence_read--
		time.Sleep(1 * time.Second)
	}

	// Після циклу перевіряємо, чи є nil
	ttl_and_sr_response, _ := Read(valueWithTtlAndsilence_read)
	fmt.Println(ttl_and_sr_response == nil)


	// with scale load
	const REQUESTS = 40_000

	// Симуляція значень для тестування
	data = map[string]interface{}{"user_id": 414243, "role": 11, "devices": map[string]string{"ios_expired_at": time.Now().String(), "android_expired_at": time.Now().String(), "mobile_app_expired_at": time.Now().String(), "external_api_integration_expired_at": time.Now().String()}, "a": 42}
	edData = map[string]interface{}{"user_id": 42, "role": 11}

	// Тестування навантаження
	fmt.Println("Checking scale load")

	for i := 1; i < 10; i++ {
		// Массив для зберігання значень, що повертаються від Q
		var values []string

		// When Q
		fmt.Println("when creates 40k tokens with Turbo Queue")
		start := time.Now()
		for i := 0; i < REQUESTS; i++ {
			values = append(values, Create(&data, nil, nil))
		}
		elapsed := time.Since(start).Seconds()
		fmt.Printf("%.3f seconds\n", elapsed)

		// When W
		fmt.Println("when reads 40k tokens")
		// index := rand.Intn(REQUESTS)
		start = time.Now()
		for i := 0; i < REQUESTS; i++ {
			Read(values[i]) // Викликаємо W з випадковим значенням
		}
		elapsed = time.Since(start).Seconds()
		fmt.Printf("%.3f seconds\n", elapsed)

		// When E
		fmt.Println("when updates 40k tokens")
		start = time.Now()
		for i := 0; i < REQUESTS; i++ {
			Update(values[i], &edData, nil, nil) // Викликаємо E
		}
		elapsed = time.Since(start).Seconds()
		fmt.Printf("%.3f seconds\n", elapsed)

		// When R
		fmt.Println("When deletes 40k tokens")
		start = time.Now()
		for i := 0; i < REQUESTS; i++ {
			Delete(values[i]) // Викликаємо R
		}
		elapsed = time.Since(start).Seconds()
		fmt.Printf("%.3f seconds\n", elapsed)
	}


	fmt.Println("when caches after read from file system")
	const LIMIT_ON_READ_FOR_CACHE = 2

	// Створення списку для зберігання попередніх значень
	var previousValues []string

	// Виконуємо Q для кількості запитів
	for i := 0; i < REQUESTS; i++ {
		previousValues = append(previousValues, Create(&data, nil, nil))
	}

	// Виконуємо ще одну серію запитів до Q
	for i := 0; i < REQUESTS; i++ {
		Create(&data, nil, nil)
	}

	// Виконуємо кешування з функцією W для попередніх значень
	for i := 0; i < LIMIT_ON_READ_FOR_CACHE; i++ {
		start := time.Now()

		for j := 0; j < REQUESTS; j++ {
			Read(previousValues[j]) // Виконуємо W для кожного попереднього значення
		}

		elapsed := time.Since(start).Seconds()
		fmt.Printf("%.3f seconds\n", elapsed)
	}
}
