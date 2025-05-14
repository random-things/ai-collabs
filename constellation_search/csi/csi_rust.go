package csi

/*
#cgo CFLAGS: -I${SRCDIR}
#cgo windows LDFLAGS: -L${SRCDIR} -lcsi
#cgo linux   LDFLAGS: -L${SRCDIR} -lcsi
#cgo darwin  LDFLAGS: -L${SRCDIR} -lcsi
#include "csi.h"
*/
import "C"

import (
	"errors"
	"unsafe"
)

type CSI struct{ h *C.CSIHandle }

func New(text string) *CSI {
	data := []byte(text)
	if len(data) == 0 {
		panic("text empty")
	}
	h := C.csi_new(
		(*C.uint8_t)(unsafe.Pointer(&data[0])),
		C.size_t(len(data)),
	)
	if h == nil {
		panic("csi_new failed")
	}
	return &CSI{h}
}

func (c *CSI) Close() {
	if c.h != nil {
		C.csi_free(c.h)
		c.h = nil
	}
}

func (c *CSI) Search(pattern string) ([]int, error) {
	pb := []byte(pattern)
	if len(pb) == 0 {
		return nil, errors.New("pattern empty")
	}
	maxOut := 1024 * 1024
	buf := make([]C.size_t, maxOut)
	n := C.csi_search(
		c.h,
		(*C.uint8_t)(unsafe.Pointer(&pb[0])), C.size_t(len(pb)),
		(*C.size_t)(unsafe.Pointer(&buf[0])), C.size_t(maxOut),
	)
	count := int(n)
	res := make([]int, count)
	for i := 0; i < count; i++ {
		res[i] = int(buf[i])
	}
	return res, nil
}
