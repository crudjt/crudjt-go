package crudjt

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
	"github.com/VladAkymov/crudjt/errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	// "crudjt/internal/grpc"
	tokenpb "github.com/VladAkymov/crudjt/proto"
  grpcserver "github.com/VladAkymov/crudjt/internal/grpc"
	"log"
	"net"
)

const CHEATCODE = "BAGUVIX" // 🐰🥚

var CacheInstance *LRUCache

type Config struct {
	EncryptedKey string
	StoreJtPath string
	WasStarted bool
	Cheatcode string
}

var config Config

func Start(cfg Config) error {
	err := ValidateEncryptedKey(cfg.EncryptedKey)
	if err != nil {
		return err
	}

	if cfg.EncryptedKey == "" {
		return fmt.Errorf(ErrorMessage(ErrorEncryptedKeyNotSet))
	}
	if cfg.WasStarted {
		return fmt.Errorf(ErrorMessage(ErrorAlreadyStarted))
	}

	cEncryptedKey := C.CString(cfg.EncryptedKey)
	defer C.free(unsafe.Pointer(cEncryptedKey))

	var cStoreJtPath *C.char
	if cfg.StoreJtPath != "" {
	    cStoreJtPath = C.CString(cfg.StoreJtPath)
	    defer C.free(unsafe.Pointer(cStoreJtPath))
	} else {
	    cStoreJtPath = nil
	}
	defer C.free(unsafe.Pointer(cStoreJtPath))

	ptr := C.start_store_jt(cEncryptedKey, cStoreJtPath)
	defer C.free(unsafe.Pointer(ptr))

	response := C.GoString(ptr)

	var res struct {
		Ok bool `json:"ok"`
		Code string `json:"code"`
		ErrorMessage string `json:"error_message"`
	}
	_ = json.Unmarshal([]byte(response), &res)

	if !res.Ok {
		if errFactory, exists := errors.ERRORS[res.Code]; exists {
			return(errFactory(res.ErrorMessage))
		}
		return(fmt.Errorf("Unknown error code %s: %s", res.Code, res.ErrorMessage))
	}

	grpc := StartGRPCServer()
	if grpc != nil {
		log.Println("Failed to start gRPC:", grpc)
	}

	cfg.WasStarted = true

	config = cfg

	if config.Cheatcode == CHEATCODE {
		fmt.Println("🐰🥚 You have activated optional param silence_read for crudjt on method Create\n" +
								"Ideal for one-time reads, email confirmation links, or limits on the number of operations\n" +
								"Each Read decrements silence_read by 1, when the counter reaches zero — the token is deleted permanently")
	}

	return nil
}

func StartGRPCServer() error {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		return err
	}

	s := grpc.NewServer()
	tokenpb.RegisterTokenServiceServer(s, &grpcserver.Server{})

	log.Println("gRPC server started on :50051")

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Printf("gRPC server stopped: %v", err)
		}
	}()

	return nil
}

var grpcClient tokenpb.TokenServiceClient

func init() {
	CacheInstance = NewLRUCache(OriginalRead)

	conn, _ := grpc.Dial(
		"localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	grpcClient = tokenpb.NewTokenServiceClient(conn)
}

func Create(hash *map[string]interface{}, ttl, silence_read *int) (string, error) {
	if !config.WasStarted {
	    return "", fmt.Errorf(ErrorMessage(ErrorNotStarted))
	}

	if config.Cheatcode != CHEATCODE {
		silence_read = nil
	}

	err := ValidateInsertion(hash, ttl, silence_read)
	if err != nil {
		return "", err
	}

  ttl_for_call := C.int64_t(-1)
  silence_read_for_call := C.int32_t(-1)

	ttl_for_cache := -1
	silence_read_for_cache := -1

  if ttl != nil {
    ttl64 := int64(*ttl)

		ttl_for_call = C.int64_t(ttl64)
		ttl_for_cache = int(*ttl) + 1
	}
	if silence_read != nil {
    silence_read32 := int32(*silence_read)

		silence_read_for_call = C.int32_t(silence_read32)
		silence_read_for_cache = int(*silence_read)
	}

	data, err := msgpack.Marshal(*hash)
	if err != nil {
		return "", err
	}

	hash_bytesize_limited := ValidateHashBytesize(len(data))
	if hash_bytesize_limited != nil {
		return "", hash_bytesize_limited
	}

	ptr := C.__create(
		(*C.uchar)(unsafe.Pointer(&data[0])), C.size_t(len(data)),
		ttl_for_call, silence_read_for_call,
	)

	token := C.GoString(ptr)
	if token == "" {
		return "", errors.NewInternalError("Something went wrong. Ups")
	}

	CacheInstance.Insert(token, *hash, ttl_for_cache, silence_read_for_cache)

	defer C.free(unsafe.Pointer(ptr))
	return token, nil
}

// func Create(hash *map[string]interface{}, ttl, silence_read *int) (string, error) {
// 	packed, err := msgpack.Marshal(data)
// 	if err != nil {
// 		return "", err
// 	}
//
// 	ttlVal := int64(-1)
// 	if ttl != nil {
// 		ttlVal = *ttl
// 	}
//
// 	silenceVal := int32(-1)
// 	if silenceW != nil {
// 		silenceVal = *silenceW
// 	}
//
// 	resp, err := client.CreateToken(context.Background(), &tokenpb.CreateTokenRequest{
// 		PackedData: packed,
// 		Ttl: ttlVal,
// 		SilenceRead: silenceVal
// 	})
//
// 	if err != nil {
// 		return "", err
// 	}
//
// 	token := resp.Token
// }

func OriginalRead(value string) {
	cValue := C.CString(value)
	defer C.free(unsafe.Pointer(cValue))

	ptr := C.__read(cValue)
	defer C.free(unsafe.Pointer(ptr))
}

func Read(value string) (map[string]interface{}, error) {
	if !config.WasStarted {
			return nil, fmt.Errorf(ErrorMessage(ErrorNotStarted))
	}

	read_err := ValidateToken(value)
	if read_err != nil {
		return nil, read_err
	}

	output := CacheInstance.Get(value)
	if output != nil {
		return output, nil
	}

	cValue := C.CString(value)
	defer C.free(unsafe.Pointer(cValue))

	ptr := C.__read(cValue)
	defer C.free(unsafe.Pointer(ptr))

	var result map[string]interface{}
	json.Unmarshal([]byte(C.GoString(ptr)), &result)

	ok, _ := result["ok"].(bool)

	if !ok {
		code, _ := result["code"].(string)
		if errFactory, exists := errors.ERRORS[code]; exists {
			return nil, errFactory(result["error_message"].(string))
		}
		return nil, fmt.Errorf("unknown error code %s", code)
	}

	if result["data"] == nil {
		return nil, nil
	}

	dataStr, ok := result["data"].(string)

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
	    return nil, fmt.Errorf("failed to parse data JSON: %w", err)
	}

	CacheInstance.ForceInsert(value, data)

	return data, nil
}

func Update(value string, hash *map[string]interface{}, ttl, silence_read *int) (bool, error) {
	if !config.WasStarted {
			return false, fmt.Errorf(ErrorMessage(ErrorNotStarted))
	}

	if config.Cheatcode != CHEATCODE {
		silence_read = nil
	}

	err := ValidateInsertion(hash, ttl, silence_read)
	if err != nil {
		return false, err
	}

  ttl_for_call := C.int64_t(-1)
  silence_read_for_call := C.int32_t(-1)

	ttl_for_cache := -1
	silence_read_for_cache := -1

  if ttl != nil {
    ttl64 := int64(*ttl)

    ttl_for_call = C.int64_t(ttl64)
		ttl_for_cache = int(*ttl) + 1
  }
  if silence_read != nil {
    silence_read32 := int32(*silence_read)

    silence_read_for_call = C.int32_t(silence_read32)
		silence_read_for_cache = int(*silence_read)
  }

  data, err := msgpack.Marshal(*hash)
  if err != nil {
    return false, err
  }

	hash_bytesize_limited := ValidateHashBytesize(len(data))
	if hash_bytesize_limited != nil {
		return false, hash_bytesize_limited
	}

	cValue := C.CString(value)
	defer C.free(unsafe.Pointer(cValue))

	ptr := C.__update(cValue, (*C.uchar)(unsafe.Pointer(&data[0])), C.size_t(len(data)), ttl_for_call, silence_read_for_call)

	bool := ptr == 1
	if bool {
		CacheInstance.Insert(value, *hash, ttl_for_cache, silence_read_for_cache)
	}

	return bool, nil
}

func Delete(value string) (bool, error) {
	if !config.WasStarted {
			return false, fmt.Errorf(ErrorMessage(ErrorNotStarted))
	}

	delete_err := ValidateToken(value)
	if delete_err != nil {
		return false, delete_err
	}

	cValue := C.CString(value)
	defer C.free(unsafe.Pointer(cValue))

	ptr := C.__delete(cValue)
	CacheInstance.Delete(value)
	return (ptr == 1), nil
}


// start gRPC server

// end gRPC server
