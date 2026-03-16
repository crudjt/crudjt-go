package main

import (
	"encoding/json"
	"fmt"
	"runtime"
	"time"
  "github.com/exwarvlad/crud_jt-go"
)

func main() {
	crudjt.StartMaster(crudjt.ServerConfig	{
	  SecretKey: "Cm7B68NWsMNNYjzMDREacmpe5sI1o0g40ZC9w1yQW3WOes7Gm59UsittLOHR2dciYiwmaYq98l3tG8h9yXVCxg==",
	  StoreJtPath: "path/to/local/storage", // optional
	  Host: "127.0.0.1", // default
	  Port: 50051, // default
	})

	fmt.Printf("OS: %s\n", runtime.GOOS)
	fmt.Printf("CPU: %s\n", runtime.GOARCH)

	// without metadata
	fmt.Println("Checking without metadata")
	data := map[string]interface{}{"user_id": 42, "role": 11}
	expectedData := map[string]interface{}{"data": data}

	edData := map[string]interface{}{"user_id": 42, "role": 8}
	expectedEdData := map[string]interface{}{"data": edData}

	value, _ := crudjt.Create(&data, nil, nil)

	result, _ := crudjt.Read(value)
	j1, _ := json.Marshal(result)
	j2, _ := json.Marshal(expectedData)
	fmt.Println(string(j1) == string(j2))

	was_updated, _ := crudjt.Update(value, &edData, nil, nil)
	fmt.Println(was_updated == true)
	result2, _ := crudjt.Read(value)

	j3, _ := json.Marshal(result2)
	j4, _ := json.Marshal(expectedEdData)
	fmt.Println(string(j3) == string(j4))

	was_deleted, _ := crudjt.Delete(value)
	fmt.Println(was_deleted == true)
	result3, _ := crudjt.Read(value)
	fmt.Println(result3 == nil)

	// with ttl
	fmt.Println("with ttl")
	ttl := 5
	valueWithttl, _ := crudjt.Create(&data, &ttl, nil)

	expectedttl := ttl
	for i := 0; i < ttl; i++ {
		expectedJSON, _ := json.Marshal(map[string]interface{}{"metadata": map[string]int{"ttl": expectedttl}, "data": data})
		result, _ := crudjt.Read(valueWithttl)
		jsonValue, _ := json.Marshal(result)
		fmt.Println(string(jsonValue) == string(expectedJSON))
		expectedttl--
		time.Sleep(1 * time.Second)
	}
	output, _ := crudjt.Read(valueWithttl)
	fmt.Println(output == nil)

	// when expired ttl
	fmt.Println("when expired ttl")
	ttl = 1
	value, _ = crudjt.Create(&data, &ttl, nil)
	time.Sleep(time.Duration(ttl) * time.Second)
	expired_ttl_output, _ := crudjt.Read(value)
	fmt.Println(expired_ttl_output == nil)

	was_updated, _ = crudjt.Update(value, &data, nil, nil)
	fmt.Println(was_updated == false)

	was_deleted, _ = crudjt.Delete(value)
	fmt.Println(was_deleted == false)

	// with silence_read
	fmt.Println("with silence_read")
	silence_read := 6
	valueWithsilence_read, _ := crudjt.Create(&data, nil, &silence_read)

	expectedsilence_read := silence_read - 1
	for i := 0; i < silence_read; i++ {
		expectedJSON, _ := json.Marshal(map[string]interface{}{"metadata": map[string]int{"silence_read": expectedsilence_read}, "data": data})
		sr_output, _ := crudjt.Read(valueWithsilence_read)
		jsonValue, _ := json.Marshal(sr_output)
		fmt.Println(string(jsonValue) == string(expectedJSON))
		expectedsilence_read--
	}
	sr_output, _ := crudjt.Read(valueWithsilence_read)
	fmt.Println(sr_output == nil)

	// with ttl and silence_read
	fmt.Println("Checking ttl and silence_read")

	data = map[string]interface{}{"user_id": 42, "role": 11}
	ttl = 5
	silence_read = ttl

	valueWithTtlAndsilence_read, _ := crudjt.Create(&data, &ttl, &silence_read)

	expectedttl = ttl
	expectedsilence_read = silence_read - 1

	for i := 0; i < silence_read; i++ {
		expectedJSON, _ := json.Marshal(map[string]interface{}{"metadata": map[string]int{"ttl": expectedttl, "silence_read": expectedsilence_read}, "data": data})
		ttl_and_sr_response, _ := crudjt.Read(valueWithTtlAndsilence_read)
		jsonValue, _ := json.Marshal(ttl_and_sr_response)

		fmt.Println(string(jsonValue) == string(expectedJSON))

		expectedttl--
		expectedsilence_read--
		time.Sleep(1 * time.Second)
	}

	ttl_and_sr_response, _ := crudjt.Read(valueWithTtlAndsilence_read)
	fmt.Println(ttl_and_sr_response == nil)

	const REQUESTS = 40_000

	data = map[string]interface{}{"user_id": 414243, "role": 11, "devices": map[string]string{"ios_expired_at": time.Now().String(), "android_expired_at": time.Now().String()}, "a": 42}
	edData = map[string]interface{}{"user_id": 42, "role": 11}

	fmt.Println("Checking scale load")

	for i := 1; i < 10; i++ {
		var values []string

		fmt.Println("when creates 40k tokens")
		start := time.Now()
		for i := 0; i < REQUESTS; i++ {
			token, _ := crudjt.Create(&data, nil, nil)
			values = append(values, token)
		}
		elapsed := time.Since(start).Seconds()
		fmt.Printf("%.3f seconds\n", elapsed)

		fmt.Println("when reads 40k tokens")
		start = time.Now()
		for i := 0; i < REQUESTS; i++ {
			crudjt.Read(values[i])
		}
		elapsed = time.Since(start).Seconds()
		fmt.Printf("%.3f seconds\n", elapsed)

		// When E
		fmt.Println("when updates 40k tokens")
		start = time.Now()
		for i := 0; i < REQUESTS; i++ {
			crudjt.Update(values[i], &edData, nil, nil)
		}
		elapsed = time.Since(start).Seconds()
		fmt.Printf("%.3f seconds\n", elapsed)

		// When R
		fmt.Println("When deletes 40k tokens")
		start = time.Now()
		for i := 0; i < REQUESTS; i++ {
			crudjt.Delete(values[i])
		}
		elapsed = time.Since(start).Seconds()
		fmt.Printf("%.3f seconds\n", elapsed)
	}


	fmt.Println("when caches after read from file system")
	const LIMIT_ON_READ_FOR_CACHE = 2

	var previousValues []string

	for i := 0; i < REQUESTS; i++ {
		token, _ := crudjt.Create(&data, nil, nil)
		previousValues = append(previousValues, token)
	}

	for i := 0; i < REQUESTS; i++ {
		crudjt.Create(&data, nil, nil)
	}

	for i := 0; i < LIMIT_ON_READ_FOR_CACHE; i++ {
		start := time.Now()

		for j := 0; j < REQUESTS; j++ {
			crudjt.Read(previousValues[j])
		}

		elapsed := time.Since(start).Seconds()
		fmt.Printf("%.3f seconds\n", elapsed)
	}
}
