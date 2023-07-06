package messageport

import (
	"encoding/json"
	"errors"
	"runtime"
	"unsafe"
)

/*
#include <CoreFoundation/CoreFoundation.h>
*/
import "C"

var (
	// ErrInvalidPort signifies that a port could not be created.
	ErrInvalidPort = errors.New("invalid port")
	// ErrSend signifies that a message could not be sent.
	ErrSend = errors.New("port send error")
)

// Remote is a remote message port.
type Remote struct {
	port C.CFMessagePortRef
}

// NewRemote creates a new remote message port.
func NewRemote(name string) (*Remote, error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	cfName := C.CFStringCreateWithCStringNoCopy(C.kCFAllocatorDefault, cName, C.kCFStringEncodingUTF8, C.kCFAllocatorNull)
	defer C.CFRelease(C.CFTypeRef(cfName))
	port := C.CFMessagePortCreateRemote(C.kCFAllocatorDefault, cfName)
	if port == 0 {
		return nil, ErrInvalidPort
	}
	p := &Remote{
		port: port,
	}
	runtime.SetFinalizer(p, func(p *Remote) {
		C.CFRelease(C.CFTypeRef(p.port))
	})
	return p, nil
}

// Close closes the message port.
func (r *Remote) Close() {
	C.CFRelease(C.CFTypeRef(r.port))
	runtime.SetFinalizer(r, nil)
}

// SendBytes sends a slice of bytes to the port.
func (r *Remote) SendBytes(id int, data []byte) (out []byte, err error) {
	cfData := C.CFDataCreateWithBytesNoCopy(C.kCFAllocatorDefault, (*C.uchar)(unsafe.Pointer(unsafe.SliceData(data))), C.long(len(data)), C.kCFAllocatorNull)
	defer C.CFRelease(C.CFTypeRef(cfData))
	var returnData C.CFDataRef
	if C.CFMessagePortSendRequest(r.port, C.int(id), cfData, 10, 10, C.kCFRunLoopDefaultMode, &returnData) == C.kCFMessagePortSuccess &&
		returnData != 0 {
		defer C.CFRelease(C.CFTypeRef(returnData))
		out, err = C.GoBytes(unsafe.Pointer(C.CFDataGetBytePtr(returnData)), C.int(C.CFDataGetLength(returnData))), nil
	} else {
		out, err = nil, ErrSend
	}
	runtime.KeepAlive(data)
	return
}

// Send sends a structure to the port.
func (r *Remote) Send(id int, in, out interface{}) (outErr error) {
	b, err := json.Marshal(in)
	if err != nil {
		return err
	}
	cfData := C.CFDataCreateWithBytesNoCopy(C.kCFAllocatorDefault, (*C.uchar)(unsafe.Pointer(unsafe.SliceData(b))), C.long(len(b)), C.kCFAllocatorNull)
	defer C.CFRelease(C.CFTypeRef(cfData))
	var returnData C.CFDataRef
	if C.CFMessagePortSendRequest(r.port, C.int(id), cfData, 10, 10, C.kCFRunLoopDefaultMode, &returnData) == C.kCFMessagePortSuccess &&
		returnData != 0 {
		defer C.CFRelease(C.CFTypeRef(returnData))
		b := unsafe.Slice((*byte)(C.CFDataGetBytePtr(returnData)), C.int(C.CFDataGetLength(returnData)))
		outErr = json.Unmarshal(b, out)
	} else {
		outErr = ErrSend
	}
	runtime.KeepAlive(b)
	return
}
