package main

/*
#include <stdlib.h>

typedef void (*EventCallback)(const char* eventJSON);

static EventCallback _event_cb = NULL;

static void setCallback(EventCallback cb) {
    _event_cb = cb;
}

static void invokeCallback(const char* json) {
    if (_event_cb != NULL) {
        _event_cb(json);
    }
}
*/
import "C"
import (
	"encoding/json"
	"sync"
	"unsafe"

	core "github.com/kushiemoon-dev/youflac-core"
)

var (
	app         *core.Core
	hasCallback bool
	mu          sync.Mutex
)

//export YouFLACInit
func YouFLACInit(dataDir *C.char) *C.char {
	mu.Lock()
	defer mu.Unlock()

	dir := C.GoString(dataDir)
	c, err := core.NewCore(dir)
	if err != nil {
		return C.CString(marshalErr("init_error", err.Error()))
	}
	app = c
	return C.CString(`{"result":"ok"}`)
}

//export YouFLACCall
func YouFLACCall(methodJSON *C.char) *C.char {
	mu.Lock()
	a := app
	mu.Unlock()

	if a == nil {
		return C.CString(marshalErr("not_initialized", "call YouFLACInit first"))
	}
	return C.CString(a.HandleRPC(C.GoString(methodJSON)))
}

//export YouFLACCallAsync
func YouFLACCallAsync(methodJSON *C.char, requestID C.int) {
	mu.Lock()
	a := app
	hasCB := hasCallback
	mu.Unlock()

	if a == nil {
		if hasCB {
			errJSON := marshalAsyncResponse(int(requestID),
				marshalErr("not_initialized", "call YouFLACInit first"))
			cs := C.CString(errJSON)
			C.invokeCallback(cs)
			C.free(unsafe.Pointer(cs))
		}
		return
	}

	input := C.GoString(methodJSON)
	rid := int(requestID)

	go func() {
		result := a.HandleRPC(input)
		payload := marshalAsyncResponse(rid, result)
		cs := C.CString(payload)
		C.invokeCallback(cs)
		C.free(unsafe.Pointer(cs))
	}()
}

//export YouFLACSetEventCallback
func YouFLACSetEventCallback(cb C.EventCallback) {
	mu.Lock()
	hasCallback = cb != nil
	mu.Unlock()

	C.setCallback(cb)

	mu.Lock()
	a := app
	mu.Unlock()

	if a != nil {
		a.SetEventCallback(func(evt core.Event) {
			data, _ := json.Marshal(evt)
			cs := C.CString(string(data))
			C.invokeCallback(cs)
			C.free(unsafe.Pointer(cs))
		})
	}
}

//export YouFLACFree
func YouFLACFree(ptr *C.char) {
	C.free(unsafe.Pointer(ptr))
}

//export YouFLACShutdown
func YouFLACShutdown() {
	mu.Lock()
	a := app
	app = nil
	mu.Unlock()

	if a != nil {
		a.Shutdown()
	}
}

func marshalErr(code, message string) string {
	resp := struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}{}
	resp.Error.Code = code
	resp.Error.Message = message
	data, _ := json.Marshal(resp)
	return string(data)
}

func marshalAsyncResponse(requestID int, payload string) string {
	resp := struct {
		RequestID int             `json:"requestId"`
		Response  json.RawMessage `json:"response"`
	}{
		RequestID: requestID,
		Response:  json.RawMessage(payload),
	}
	data, _ := json.Marshal(resp)
	return string(data)
}

func main() {} // required for c-shared/c-archive
