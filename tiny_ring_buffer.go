package tiny_ring_bufio

import (
	"errors"
	"fmt"
	"io"
)

const (
	DEFAULT_BUF_SIZE = 4096 * 2
)

type TinyRBuff struct {
	Buf     []byte
	Head    int
	Tail    int
	Checked int
	min     int
	OutHead int
	DupSize int
	ReadMax int
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
		min:     min,
		ReadMax: 4096,
	}
}

func (t *TinyRBuff) All() []byte {
	return t.Buf[t.Head:len(t.Buf)]
}
func (t *TinyRBuff) Len() int {
	return len(t.Buf)
}

func (t *TinyRBuff) ReadAtLeast(r io.Reader, must int) (size int, err error) {
	//check me: should use readv ?
	end := 0
	if t.Head < t.Tail {
		end = t.Tail
	} else if t.Head > t.Tail && t.OutHead > 0 {
		return 0, nil
	} else {
		end = len(t.Buf) - t.min
	}
	if end-t.Head > t.ReadMax {
		end = t.Head + t.ReadMax
	}
	size, err = io.ReadAtLeast(r, t.Buf[t.Head:end], must)
	t.Head += size
	if t.Head >= len(t.Buf)-t.min {
		t.OutHead = t.Head
		t.Head = len(t.Buf) - t.Head
	}
	return size, err
}
func (t *TinyRBuff) Use() []byte {
	if t.Tail < t.Checked {
		old_tail := t.Tail
		t.Tail = t.Checked
		return t.Buf[old_tail:t.Checked]
	}
	if t.Tail < t.OutHead {
		old_tail := t.Tail
		out_head := t.OutHead
		t.Tail = out_head
		t.OutHead = 0
		return t.Buf[old_tail:out_head]
	}
	return t.Buf[0:0]
}

func (t *TinyRBuff) WriteAt(w io.WriterAt, size int) (w_len int, err error) {
	//check me: should use writev
	if t.Tail <= t.Checked {
		w_len, err = w.WriteAt(t.Buf[t.Tail:t.Checked], int64(size))
		t.Tail += w_len
		size += w_len
		return w_len, err
	}
	if t.Tail < t.OutHead {
		w_len, err = w.WriteAt(t.Buf[t.Tail:t.OutHead], int64(t.OutHead-t.Tail))
		t.Tail += w_len
		if t.Tail == t.OutHead {
			t.Tail = t.DupSize
			t.OutHead = 0
			t.DupSize = 0
		}
		if err != nil || t.OutHead != 0 {
			return w_len, err
		}
		var out_len int
		out_len, err = w.WriteAt(t.Buf[t.Tail:t.Checked], int64(size-w_len))
		w_len += out_len
		t.Tail += out_len

		return w_len, err
	}
	return 0, errors.New("this data is not written")
}

func (t *TinyRBuff) UnCheckedSeqLen() int {
	check_tail := func(size int) int {
		if t.Tail < t.min && t.OutHead > 0 {
			return size - 1
		}
		return size
	}

	if t.OutHead > t.Checked {
		if t.OutHead-t.Checked < t.min {
			if t.OutHead-t.Checked+t.Head > t.min {
				copy(t.Buf[t.OutHead:t.Checked+t.min], t.Buf[0:t.Checked+t.min-t.OutHead])
				t.DupSize = t.Checked + t.min - t.OutHead
				t.OutHead = t.Checked + t.min
			}
		}
		return check_tail(t.OutHead - t.Checked)

	} else if t.Head > t.Checked {
		return check_tail(t.Head - t.Checked)
	}
	return 0
}
func (t *TinyRBuff) UnCheckedLen() int {
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
}
func (t *TinyRBuff) AllLen() int {
	if t.Tail <= t.Checked {
		return t.Checked - t.Tail
	}
	if t.Tail < t.OutHead {
		return t.OutHead - t.Tail + t.Checked - t.DupSize
	}
	//FIXME: should not falldown
	return 0

}

func (t *TinyRBuff) Check(size int) []byte {
	old_check := t.Checked
	t.Checked += size
	fmt.Println(t.p())
	if t.Checked >= t.OutHead {
		t.Checked = t.Checked - t.OutHead + t.DupSize
		t.OutHead = 0
	}
	return t.Buf[old_check : old_check+size]
}

func (t *TinyRBuff) Checkv(size int) {
	if t.Checked < t.Head {
		t.Check(t.Head - t.Checked)
		return
	}
	diff := t.OutHead - t.Checked
	t.Check(diff)
	t.Check(size - diff)
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
	return t.Checked - t.Tail
}
