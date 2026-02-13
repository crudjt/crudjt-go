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
	"net"
	"context"
)

var CacheInstance *LRUCache

type mode int

const (
	modeNone mode = iota
	modeMaster
	modeClient
)

type ServerConfig struct {
	EncryptedKey string
	StoreJtPath string
	Host string
	Port int
}

type grpcServer struct {
	tokenpb.UnimplementedTokenServiceServer
}

type ClientConfig struct {
	Address string
}

type runtimeState struct {
	mode mode
	grpcServer *grpc.Server
	grpcClient tokenpb.TokenServiceClient
}

var state runtimeState

func StartMaster(cfg ServerConfig) error {
	if state.mode != modeNone {
		return fmt.Errorf("CRUDJT already initialized")
	}

	err := ValidateEncryptedKey(cfg.EncryptedKey)
	if err != nil {
		return err
	}

	if cfg.EncryptedKey == "" {
		return fmt.Errorf(ErrorMessage(ErrorEncryptedKeyNotSet))
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

	server, err := StartGRPCServer(cfg)
	if err != nil {
		return err
	}

	state.grpcServer = server
	state.mode = modeMaster

	// cfg.WasStarted = true

	// config = cfg

	return nil
}

func StartGRPCServer(cfg ServerConfig) (*grpc.Server, error) {
	address := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	lis, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}

	server := grpc.NewServer()
	tokenpb.RegisterTokenServiceServer(server, &grpcServer{})

	go server.Serve(lis)

	return server, nil
}

var grpcClient tokenpb.TokenServiceClient

func init() {
	CacheInstance = NewLRUCache(ReadNif)

	conn, _ := grpc.Dial(
		"localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	grpcClient = tokenpb.NewTokenServiceClient(conn)
}

func OriginalCreate(hash *map[string]interface{}, ttl, silence_read *int) (string, error) {
	if state.mode != modeNone {
		return "", fmt.Errorf("CRUDJT already initialized")
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

func Create(hash *map[string]interface{}, ttl, silence_read *int) (string, error) {
	packed, err := msgpack.Marshal(*hash)
		if err != nil {
			return "", err
		}

		req := &tokenpb.CreateTokenRequest{
			PackedData: packed,
			Ttl: intPtrToInt64(ttl),
			SilenceRead: intPtrToInt64(silence_read),
		}

		resp, err := grpcClient.CreateToken(context.Background(), req)
		// fmt.Println(resp)
		// fmt.Println(err)
		if err != nil {
			return "", err
		}

		return resp.Token, nil
}

func ReadNif(value string) {
	cValue := C.CString(value)
	defer C.free(unsafe.Pointer(cValue))

	ptr := C.__read(cValue)
	defer C.free(unsafe.Pointer(ptr))
}

func OriginalRead(value string) (map[string]interface{}, error) {
	if state.mode != modeNone {
		return nil, fmt.Errorf("CRUDJT already initialized")
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

func Read(value string) (map[string]interface{}, error) {
	req := &tokenpb.ReadTokenRequest{
			Token: value,
		}

	resp, err := grpcClient.ReadToken(context.Background(), req)
	if err != nil {
		return nil, err
	}

	data := map[string]any{}
	if err := msgpack.Unmarshal(resp.PackedData, &data); err != nil {
		return nil, err
	}

	return data, nil
}

func OriginalUpdate(value string, hash *map[string]interface{}, ttl, silence_read *int) (bool, error) {
	if state.mode != modeNone {
		return false, fmt.Errorf("CRUDJT already initialized")
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

func Update(value string, hash *map[string]interface{}, ttl, silence_read *int) (bool, error) {
	packed, err := msgpack.Marshal(hash)
	if err != nil {
		return false, err
	}

	req := &tokenpb.UpdateTokenRequest{
		Token: value,
		PackedData: packed,
		Ttl: intPtrToInt64(ttl),
		SilenceRead: intPtrToInt64(silence_read),
	}

	resp, err := grpcClient.UpdateToken(context.Background(), req)
	if err != nil {
		return false, err
	}

	return resp.Result, nil
}

func OriginalDelete(value string) (bool, error) {
	if state.mode != modeNone {
		return false, fmt.Errorf("CRUDJT already initialized")
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

func Delete(value string) (bool, error) {
	req := &tokenpb.DeleteTokenRequest{
		Token: value,
	}

	resp, err := grpcClient.DeleteToken(context.Background(), req)
	if err != nil {
		return false, err
	}

	return resp.Result, nil
}

// start gRPC server
func (s *grpcServer) CreateToken(
	ctx context.Context,
	req *tokenpb.CreateTokenRequest,
) (*tokenpb.CreateTokenResponse, error) {

	data := map[string]any{}
	if err := msgpack.Unmarshal(req.PackedData, &data); err != nil {
		return nil, err
	}

	var ttlInt *int
	if req.Ttl != -1 {
		v := int(req.Ttl)
		ttlInt = &v
	}

	var silenceInt *int
	if req.SilenceRead != -1 {
		v := int(req.SilenceRead)
		silenceInt = &v
	}

	token, err := OriginalCreate(&data, ttlInt, silenceInt)
	if err != nil {
		return nil, err
	}

	return &tokenpb.CreateTokenResponse{Token: token}, nil
}

func ReadToken(token string) (map[string]any, error) {
	req := &tokenpb.ReadTokenRequest{
		Token: token,
	}

	resp, err := grpcClient.ReadToken(context.Background(), req)
	if err != nil {
		return nil, err
	}

	data := map[string]any{}
	if err := msgpack.Unmarshal(resp.PackedData, &data); err != nil {
		return nil, err
	}

	return data, nil
}

func UpdateToken(token string, data map[string]any, ttl, silenceW *int) (bool, error) {
	packed, err := msgpack.Marshal(data)
	if err != nil {
		return false, err
	}

	req := &tokenpb.UpdateTokenRequest{
		Token: token,
		PackedData: packed,
		Ttl: intPtrToInt64(ttl),
		SilenceRead: intPtrToInt64(silenceW),
	}

	resp, err := grpcClient.UpdateToken(context.Background(), req)
	if err != nil {
		return false, err
	}

	return resp.Result, nil
}

func DeleteToken(token string) (bool, error) {
	req := &tokenpb.DeleteTokenRequest{
		Token: token,
	}

	resp, err := grpcClient.DeleteToken(context.Background(), req)
	if err != nil {
		return false, err
	}

	return resp.Result, nil
}
// end gRPC server

func intPtrToInt64(v *int) int64 {
	if v == nil {
		return -1 // використовуємо -1 як "nil" для proto
	}
	return int64(*v)
}
