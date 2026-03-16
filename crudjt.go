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
	"github.com/exwarvlad/crud_jt-go/errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	tokenpb "github.com/exwarvlad/crud_jt-go/proto"
	"net"
	"context"
	"strconv"
)

var CacheInstance *LRUCache

type mode int

const (
	modeNone mode = iota
	modeMaster
	modeClient
)

const (
	DefaultHost = "127.0.0.1"
	DefaultPort = 50051
)

type ServerConfig struct {
	SecretKey string
	StoreJtPath string
	Host string
	Port int
}

type grpcServer struct {
	tokenpb.UnimplementedTokenServiceServer
}

type ClientConfig struct {
	Host string
	Port int
}

type runtimeState struct {
	mode mode
	grpcServer *grpc.Server
	grpcClient tokenpb.TokenServiceClient
}

var state runtimeState

func StartMaster(cfg ServerConfig) error {
	if state.mode != modeNone {
		return fmt.Errorf(ErrorMessage(ErrorAlreadyStarted))
	}

	err := ValidateSecretKey(cfg.SecretKey)
	if err != nil {
		return err
	}

	if cfg.SecretKey == "" {
		return fmt.Errorf(ErrorMessage(ErrorSecretKeyNotSet))
	}

	host := cfg.Host
	if host == "" {
		host = DefaultHost
	}

	port := cfg.Port
	if port == 0 {
		port = DefaultPort
	}

	cSecretKey := C.CString(cfg.SecretKey)
	defer C.free(unsafe.Pointer(cSecretKey))

	var cStoreJtPath *C.char
	if cfg.StoreJtPath != "" {
	    cStoreJtPath = C.CString(cfg.StoreJtPath)
	    defer C.free(unsafe.Pointer(cStoreJtPath))
	}

	ptr := C.start_store_jt(cSecretKey, cStoreJtPath)
	if ptr == nil {
	    return fmt.Errorf("start_store_jt returned nil")
	}
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

	address := fmt.Sprintf("%s:%d", host, port)
	server, err := StartGRPCServer(address)
	if err != nil {
		return err
	}

	state.grpcServer = server
	state.mode = modeMaster

	return nil
}

func ConnectToMaster(cfg ClientConfig) error {
	host := cfg.Host
	if host == "" {
		host = DefaultHost
	}

	port := cfg.Port
	if port == 0 {
		port = DefaultPort
	}

	address := net.JoinHostPort(host, strconv.Itoa(port))

	conn, err := grpc.Dial(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return err
	}

	state.grpcClient = tokenpb.NewTokenServiceClient(conn)
	state.mode = modeClient

	return nil
}

func StartGRPCServer(address string) (*grpc.Server, error) {
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
}

func OriginalCreate(hash *map[string]interface{}, ttl, silence_read *int) (string, error) {
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
	switch state.mode {

	case modeMaster:
		return OriginalCreate(hash, ttl, silence_read)

	case modeClient:
		packed, err := msgpack.Marshal(*hash)
			if err != nil {
				return "", err
			}

			req := &tokenpb.CreateTokenRequest{
				PackedData: packed,
				Ttl: intPtrToInt64(ttl),
				SilenceRead: intPtrToInt64(silence_read),
			}

			resp, err := state.grpcClient.CreateToken(context.Background(), req)
			if err != nil {
				return "", err
			}

			return resp.Token, nil

	default:
		return "", fmt.Errorf(ErrorMessage(ErrorAlreadyStarted))
	}
}

func ReadNif(value string) {
	cValue := C.CString(value)
	defer C.free(unsafe.Pointer(cValue))

	ptr := C.__read(cValue)
	defer C.free(unsafe.Pointer(ptr))
}

func OriginalRead(value string) (map[string]interface{}, error) {
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
	switch state.mode {

	case modeMaster:
		return OriginalRead(value)

	case modeClient:
		req := &tokenpb.ReadTokenRequest{
				Token: value,
			}

		resp, err := state.grpcClient.ReadToken(context.Background(), req)
		if err != nil {
			return nil, err
		}

		data := map[string]any{}
		if err := msgpack.Unmarshal(resp.PackedData, &data); err != nil {
			return nil, err
		}

		return data, nil

	default:
		return nil, fmt.Errorf(ErrorMessage(ErrorAlreadyStarted))
	}
}

func OriginalUpdate(value string, hash *map[string]interface{}, ttl, silence_read *int) (bool, error) {
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
	switch state.mode {

	case modeMaster:
		return OriginalUpdate(value, hash, ttl, silence_read)

	case modeClient:
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

		resp, err := state.grpcClient.UpdateToken(context.Background(), req)
		if err != nil {
			return false, err
		}

		return resp.Result, nil

	default:
		return false, fmt.Errorf(ErrorMessage(ErrorAlreadyStarted))
	}
}

func OriginalDelete(value string) (bool, error) {
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
	switch state.mode {

	case modeMaster:
		return OriginalDelete(value)

	case modeClient:
		req := &tokenpb.DeleteTokenRequest{
			Token: value,
		}

		resp, err := state.grpcClient.DeleteToken(context.Background(), req)
		if err != nil {
			return false, err
		}

		return resp.Result, nil

	default:
		return false, fmt.Errorf(ErrorMessage(ErrorAlreadyStarted))
	}
}

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

func (s *grpcServer) ReadToken(
	ctx context.Context,
	req *tokenpb.ReadTokenRequest,
) (*tokenpb.ReadTokenResponse, error) {

	rawToken := req.GetToken()

	resultHash, err := OriginalRead(rawToken)
	if err != nil {
		return nil, err
	}

	packedData, err := msgpack.Marshal(resultHash)
	if err != nil {
		return nil, err
	}

	return &tokenpb.ReadTokenResponse{
		PackedData: packedData,
	}, nil
}

func (s *grpcServer) UpdateToken(
	ctx context.Context,
	req *tokenpb.UpdateTokenRequest,
) (*tokenpb.UpdateTokenResponse, error) {

	rawToken := req.GetToken()

	packedData := req.GetPackedData()

	unpacked := make(map[string]interface{})
	if err := msgpack.Unmarshal(packedData, &unpacked); err != nil {
		return nil, err
	}

	var ttlPtr *int
	if req.GetTtl() != -1 {
		t := int(req.GetTtl())
		ttlPtr = &t
	}

	var silenceWPtr *int
	if req.GetSilenceRead() != -1 {
		sw := int(req.GetSilenceRead())
		silenceWPtr = &sw
	}

	result, err := OriginalUpdate(rawToken, &unpacked, ttlPtr, silenceWPtr)
	if err != nil {
		return nil, err
	}

	return &tokenpb.UpdateTokenResponse{
		Result: result,
	}, nil
}

func (s *grpcServer) DeleteToken(
	ctx context.Context,
	req *tokenpb.DeleteTokenRequest,
) (*tokenpb.DeleteTokenResponse, error) {

	rawToken := req.GetToken()

	result, err := OriginalDelete(rawToken)
	if err != nil {
		return nil, err
	}

	return &tokenpb.DeleteTokenResponse{
		Result: result,
	}, nil
}

// end gRPC server

func intPtrToInt64(v *int) int64 {
	if v == nil {
		return -1
	}
	return int64(*v)
}
