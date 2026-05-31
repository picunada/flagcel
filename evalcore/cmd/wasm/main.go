package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"sync"
	"unsafe"

	"github.com/picunada/flagcel/evalcore"
)

const resultHeaderSize = 8

var (
	mu            sync.Mutex
	allocations   = map[uint32][]byte{}
	evaluator     *evalcore.Evaluator
	version       uint64
	lastLoadError string
)

func main() {}

//go:wasmexport alloc
func alloc(size uint32) uint32 {
	if size == 0 {
		size = 1
	}
	buf := make([]byte, size)
	ptr := ptrOf(buf)
	if ptr == 0 {
		return 0
	}

	mu.Lock()
	allocations[ptr] = buf
	mu.Unlock()
	return ptr
}

//go:wasmexport free
func free(ptr uint32, size uint32) {
	mu.Lock()
	delete(allocations, ptr)
	mu.Unlock()
}

//go:wasmexport loadDefinitions
func loadDefinitions(ptr uint32, length uint32) uint64 {
	data := readBytes(ptr, length)
	if data == nil && length > 0 {
		setLoadError("invalid definitions pointer")
		return 0
	}

	var defs evalcore.Definitions
	if err := decodeJSON(data, &defs); err != nil {
		setLoadError(err.Error())
		return 0
	}

	loaded, err := evalcore.Load(defs)
	if err != nil {
		setLoadError(err.Error())
		return 0
	}

	mu.Lock()
	evaluator = loaded
	version++
	current := version
	lastLoadError = ""
	mu.Unlock()
	return current
}

//go:wasmexport evaluate
func evaluate(keyPtr uint32, keyLen uint32, ctxPtr uint32, ctxLen uint32) uint32 {
	keyBytes := readBytes(keyPtr, keyLen)
	if keyBytes == nil && keyLen > 0 {
		return writeResult(errorResult("invalid key pointer"))
	}

	ctx, err := readContext(ctxPtr, ctxLen)
	if err != nil {
		return writeResult(errorResult(err.Error()))
	}

	mu.Lock()
	loaded := evaluator
	loadErr := lastLoadError
	mu.Unlock()
	if loaded == nil {
		if loadErr == "" {
			loadErr = "definitions not loaded"
		}
		return writeResult(errorResult(loadErr))
	}

	result := loaded.Evaluate(string(keyBytes), ctx)
	return writeResult(result)
}

//go:wasmexport evaluateAll
func evaluateAll(ctxPtr uint32, ctxLen uint32) uint32 {
	ctx, err := readContext(ctxPtr, ctxLen)
	if err != nil {
		return writeResult(errorResult(err.Error()))
	}

	mu.Lock()
	loaded := evaluator
	loadErr := lastLoadError
	mu.Unlock()
	if loaded == nil {
		if loadErr == "" {
			loadErr = "definitions not loaded"
		}
		return writeResult(map[string]evalcore.EvaluationResult{
			"": errorResult(loadErr),
		})
	}

	return writeResult(loaded.EvaluateAll(ctx))
}

func setLoadError(message string) {
	mu.Lock()
	evaluator = nil
	lastLoadError = message
	mu.Unlock()
}

func readContext(ptr uint32, length uint32) (evalcore.DataContext, error) {
	data := readBytes(ptr, length)
	if data == nil && length > 0 {
		return nil, jsonError("invalid context pointer")
	}
	if len(data) == 0 {
		return evalcore.DataContext{}, nil
	}

	var ctx map[string]any
	if err := decodeJSON(data, &ctx); err != nil {
		return nil, err
	}
	return evalcore.DataContext(ctx), nil
}

func decodeJSON(data []byte, out any) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	return dec.Decode(out)
}

func writeResult(value any) uint32 {
	payload, err := json.Marshal(value)
	if err != nil {
		payload, _ = json.Marshal(errorResult(err.Error()))
	}

	buf := make([]byte, resultHeaderSize+len(payload))
	binary.LittleEndian.PutUint64(buf[:resultHeaderSize], uint64(len(payload)))
	copy(buf[resultHeaderSize:], payload)

	ptr := ptrOf(buf)
	if ptr == 0 {
		return 0
	}
	mu.Lock()
	allocations[ptr] = buf
	mu.Unlock()
	return ptr
}

func errorResult(message string) evalcore.EvaluationResult {
	return evalcore.EvaluationResult{
		Value:     false,
		ValueType: evalcore.ValueTypeBoolean,
		Reason:    "error",
		Error:     message,
	}
}

func jsonError(message string) error {
	return errors.New(message)
}

func readBytes(ptr uint32, length uint32) []byte {
	if length == 0 {
		return nil
	}

	mu.Lock()
	if buf, ok := allocations[ptr]; ok && uint32(len(buf)) >= length {
		out := append([]byte(nil), buf[:length]...)
		mu.Unlock()
		return out
	}
	mu.Unlock()

	return append([]byte(nil), unsafe.Slice((*byte)(unsafe.Pointer(uintptr(ptr))), length)...)
}

func ptrOf(buf []byte) uint32 {
	if len(buf) == 0 {
		return 0
	}
	return uint32(uintptr(unsafe.Pointer(&buf[0])))
}
