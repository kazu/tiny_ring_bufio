package tiny_ring_bufio

import (
	"errors"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
)

const (
	DEFAULT_BUF_SIZE = 4096 * 2
)

type TinyRBuff struct {
	Buf     []byte
	Head    uint64
	Tail    uint64
	Checked uint64
	min     uint64
	OutHead uint64
	DupSize uint64
	ReadMax uint64

	MuR sync.Mutex
}

func (t *TinyRBuff) P() string {
	return t.p()
}
func (t *TinyRBuff) p() string {
	return fmt.Sprintf("<Head:%d, Tail:%d, Checked:%d, min: %d, OutHead: %d, DupSize: %d ReadMax: %d>",
		t.Head, t.Tail, t.Checked, t.min, t.OutHead, t.DupSize, t.ReadMax)
}

func New(size int, min int) *TinyRBuff {
	return NewTinyRBuff(size, min)
}
func NewTinyRBuff(size int, min int) *TinyRBuff {
	return &TinyRBuff{
		Buf:     make([]byte, size),
		min:     uint64(min),
		ReadMax: 4096,
	}
}

func (t *TinyRBuff) All() []byte {
	return t.Buf[t.Head:len(t.Buf)]
}
func (t *TinyRBuff) Len() int {
	return len(t.Buf)
}

func (t *TinyRBuff) ReadAtLeast(r io.Reader, omust int) (size int, err error) {
	//check me: should use readv ?
	end := uint64(0)
	must := uint64(omust)
	defer func() {
		//		fmt.Println("bufio.ReadAtLeast, must, return size, len(buf) ", must, size, len(t.Buf), t.P())
	}()
	if t.Head < t.Tail {
		end = t.Tail
	} else if t.Head > t.Tail && t.OutHead > 0 {
		return 0, nil
	} else {
		end = uint64(len(t.Buf)) - t.min
	}
	if end-t.Head > t.ReadMax {
		end = t.Head + t.ReadMax
	}
	if end-t.Head < must {
		must = end - t.Head
	}
	t.MuR.Lock()
	size, err = io.ReadAtLeast(r, t.Buf[int(t.Head):int(end)], int(must))
	t.MuR.Unlock()
RETRY:
	if !atomic.CompareAndSwapUint64(&t.Head, t.Head, t.Head+uint64(size)) {
		goto RETRY
	}
	//t.Head += uint64(size)
	if t.Head >= uint64(len(t.Buf))-t.min {
		t.OutHead = t.Head
		t.Head = t.Head - uint64(len(t.Buf)) + t.min
		t.DupSize = 0
	}
	return size, err
}
func (t *TinyRBuff) Use() []byte {
	if t.Tail < t.Checked {
		var old_tail uint64
		for {
			old_tail = t.Tail
			if atomic.CompareAndSwapUint64(&t.Tail, t.Tail, t.Checked) {
				break
			}
		}
		//t.Tail = t.Checked
		return t.Buf[old_tail:t.Checked]
	}
	if t.Tail < t.OutHead {
		var old_tail, out_head uint64
		for {
			old_tail = t.Tail
			out_head = t.OutHead
			t.Tail = out_head
			if atomic.CompareAndSwapUint64(&t.Tail, t.Tail, out_head) {
				break
			}
		}
		t.OutHead = 0
		return t.Buf[old_tail:out_head]
	}
	return t.Buf[0:0]
}

func (t *TinyRBuff) WriteAt(w io.WriterAt, off int) (w_len int, err error) {
	//check me: should use writev

	//w_len := uint64(ow_len)
	defer func() {
		//		fmt.Println("bufio.Write, off ", off, t.P())
	}()

	if t.Tail <= t.Checked {
		w_len, err = w.WriteAt(t.Buf[int(t.Tail):int(t.Checked)], int64(off))
		for {
			if atomic.CompareAndSwapUint64(&t.Tail, t.Tail, t.Tail+uint64(w_len)) {
				break
			}
			//t.Tail += uint64(w_len)
		}
		off += w_len
		return w_len, err
	}
	if t.Tail < t.OutHead {

		w_len, err = w.WriteAt(t.Buf[t.Tail:t.OutHead], int64(off))
		for {
			if atomic.CompareAndSwapUint64(&t.Tail, t.Tail, t.Tail+uint64(w_len)) {
				break
			}
			//t.Tail += uint64(w_len)
		}
		//t.Tail += uint64(w_len)
		if t.Tail == t.OutHead {
			t.Tail = 0
			t.OutHead = 0
			t.DupSize = 0
		}
		if err != nil || t.OutHead != 0 {
			return w_len, err
		}
		var out_len int
		out_len, err = w.WriteAt(t.Buf[t.Tail:t.Checked], int64(off+w_len))
		w_len += out_len
		for {
			if atomic.CompareAndSwapUint64(&t.Tail, t.Tail, t.Tail+uint64(out_len)) {
				break
			}
			//t.Tail += uint64(w_len)
		}
		//w_len += out_len
		//t.Tail += uint64(out_len)

		return w_len, err
	}
	return 0, errors.New("this data is not written")
}

func (t *TinyRBuff) UnCheckedSeqLen() int {
	check_tail := func(size uint64) int {
		if size > 0 && t.Tail < t.min && t.OutHead > 0 {
			return int(size - uint64(1))
		}
		return int(size)
	}
	if t.Head == t.Checked && t.Checked == t.Tail && t.OutHead > 0 {
		return check_tail(t.OutHead - t.Checked)
	} else if t.Head >= t.Checked {
		return check_tail(t.Head - t.Checked)
	} else if t.OutHead > t.Checked {
		if t.OutHead-t.Checked < t.min {
			if t.OutHead-t.Checked+t.Head >= t.min {
				copy(t.Buf[t.OutHead:t.Checked+t.min], t.Buf[0:t.Checked+t.min-t.OutHead])
				for {
					if atomic.CompareAndSwapUint64(&t.DupSize, t.DupSize, t.Checked+t.min-t.OutHead) {
						if atomic.CompareAndSwapUint64(&t.OutHead, t.OutHead, uint64(len(t.Buf))-t.min) {
							break
						}
					}
				}
				//t.DupSize = t.Checked + t.min - t.OutHead
				//t.OutHead = uint64(len(t.Buf)) - t.min
			}
		}
		return check_tail(t.OutHead - t.Checked + t.DupSize)
	}
	return 0
}
func (t *TinyRBuff) UnCheckedLen() int {
	if t.OutHead == 0 || t.Head >= t.Checked {
		return t.UnCheckedSeqLen()
	} else {
		return t.UnCheckedSeqLen() + int(t.Head-t.DupSize)
	}
	/*
		if t.Tail < t.min && t.OutHead > 0 {
			return t.UnCheckedSeqLen()
		}
		if t.Checked < t.OutHead {
			return t.OutHead - t.Checked - t.DupSize + t.Head
		}
		if t.Head > t.Checked {
			return t.Head - t.Checked
		}

		return 0
	*/
}
func (t *TinyRBuff) AllLen() int {
	if t.Tail <= t.Checked {
		return int(t.Checked - t.Tail)
	}
	if t.Tail < t.OutHead {
		return int(t.OutHead - t.Tail + t.Checked - t.DupSize)
	}
	//FIXME: should not falldown
	return 0

}

func (t *TinyRBuff) SeqMin() int {
	return int(t.min)
}
func (t *TinyRBuff) Check(size int) []byte {
	defer func() {
		//		fmt.Println("bufio.Check, size ", size, t.P())
	}()

	var old_check uint64
	for {
		old_check = t.Checked
		if atomic.CompareAndSwapUint64(&t.Checked, t.Checked, t.Checked+uint64(size)) {
			break
		}
		//t.Checked += uint64(size)

	}
	//old_check := t.Checked
	//t.Checked += uint64(size)
	if t.Checked >= t.OutHead {
		for {
			if atomic.CompareAndSwapUint64(&t.Checked, t.Checked, t.Checked-t.OutHead) {
				break
			}
		}
		//t.Checked = t.Checked - t.OutHead
	}
	if old_check > uint64(len(t.Buf)) || old_check+uint64(size) > uint64(len(t.Buf)) {
		fmt.Printf("WARN: bufio overrun buf_len=%d offset=%d size=%d", len(t.Buf), old_check, size)
		old_check = uint64(len(t.Buf) - size)
	}

	return t.Buf[old_check : int(old_check)+int(size)]
}

func (t *TinyRBuff) Checkv(size int) {
	if t.Checked < t.Head {
		t.Check(int(t.Head - t.Checked))
		return
	}
	diff := t.OutHead - t.Checked
	if diff >= uint64(size) {
		t.Check(size)
	} else {
		t.Check(int(diff))
		t.Check(size - int(diff))
	}
	return
}

func (t *TinyRBuff) HasCheckedBuf() bool {
	if t.Tail < t.Checked {
		return true
	}
	return false
}
func (t *TinyRBuff) CheckedBuf() []byte {
	return t.Buf[t.Tail:t.Checked]
}
func (t *TinyRBuff) CheckedLen() int {
	return int(t.Checked - t.Tail)
}
