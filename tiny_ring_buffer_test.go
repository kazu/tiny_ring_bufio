package tiny_ring_bufio

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func CreateFile(size int) {

	file, _ := os.Create("trbtest")
	for i := 0; i < size; i++ {
		file.Write(([]byte)(strconv.Itoa(i)))
	}
	file.Close()
}

func BufFile(size int) io.Reader {

	file := bytes.NewBuffer(nil)

	for i := 0; i < size; i++ {
		d := i % 10
		file.Write(([]byte)(strconv.Itoa(d)))
	}
	return file
}

func TestReadAtLeast9000(t *testing.T) {

	file := BufFile(90000)

	bufio := NewTinyRBuff(4096*2, 2)
	fmt.Printf("buf=%s\n", bufio.P())
	file_n, e := bufio.ReadAtLeast(file, 2)

	assert.NoError(t, e)
	assert.True(t, file_n > 0)

	bufio.Check(2)
	bufio.Check(4000)
	bufio.Use()

	file_n, e = bufio.ReadAtLeast(file, 500)

	fmt.Printf("buf=%s\n", bufio.P())
	fmt.Printf("UnCheckedSeqLen=%d\n", bufio.UnCheckedSeqLen())
	file_n, e = bufio.ReadAtLeast(file, 500)
	bufio.Check(4188)
	bufio.Use()
	assert.Equal(t, bufio.Tail, uint64(0))

	bufio.Check(500)
	bufio.Use()
	assert.NoError(t, e)
}

func TestReadAtLeast(t *testing.T) {

	CreateFile(4000)
	file, _ := os.Open("trbtest")

	bufio := NewTinyRBuff(4096*2, 20)
	file_n, e := bufio.ReadAtLeast(file, 20)

	if e != nil && bufio.UnCheckedSeqLen() < 1 {
		t.Error("ReadAtLeast", e, bufio.UnCheckedSeqLen())
	}
	fmt.Println("bufio fine_n", bufio.p(), file_n)
	fmt.Println("unchecked seq len", bufio.UnCheckedSeqLen(), bufio.UnCheckedLen())

	data := bufio.Check(bufio.UnCheckedSeqLen())

	fmt.Println("file_n, buf_io", file_n, bufio.p())
	//return
	if data[20] != ([]byte)("1")[0] {
		t.Error("data[3] != ", ([]byte)("4"))
	}
	file.Close()
	return
	fmt.Println("bufio fine_n", bufio.p(), file_n)
	fmt.Println("unchecked seq len", bufio.UnCheckedSeqLen(), bufio.UnCheckedLen())

}
func TestWriteAt(t *testing.T) {
	CreateFile(4000)
	file, _ := os.Open("trbtest")
	outfile, _ := os.Create("outtest")

	bufio := NewTinyRBuff(4096*2, 20)
	bufio.ReadMax = 4096 * 2
	_, e := bufio.ReadAtLeast(file, 20)
	fmt.Println("!after read bufio", bufio.P())

	if e != nil || bufio.UnCheckedSeqLen() < 1 {
		t.Error("ReadAtLeast fail", e, bufio.UnCheckedSeqLen(), bufio.P())
	}
	bufio.Check(bufio.UnCheckedSeqLen())
	fmt.Println("after check bufio", bufio.p())

	written, _ := bufio.WriteAt(outfile, 0)
	fmt.Println("bufio written", bufio.p(), written)
	fmt.Println("UnCeckedLne()", bufio.UnCheckedLen())
	bufio.Check(bufio.UnCheckedLen())
	fmt.Println("recheck bufio ", bufio.p())
	outfile.Close()

	file, _ = os.Open("outtest")
	data := make([]byte, 30)
	file.Read(data[0:30])

	if data[20] != ([]byte)("1")[0] {
		t.Error("data[3] != ", ([]byte)("4"))
	}
	file.Close()
}
func TestResetHead(t *testing.T) {
	CreateFile(4000)

	file, _ := os.Open("trbtest")

	bufio := NewTinyRBuff(4096*2, 20)
	bufio.ReadMax = 4096
	bufio.Head = 8103
	bufio.Tail = 8103
	bufio.OutHead = 0
	bufio.DupSize = 0
	bufio.Checked = 8103

	bufio.ReadAtLeast(file, 20)

	file.Close()
	if bufio.Head != 0 {
		t.Error("bufio.Head is not reset", bufio.P())
	}

}
func TestResetCheck(t *testing.T) {
	CreateFile(4000)

	bufio := NewTinyRBuff(4096*2, 20)
	bufio.ReadMax = 4096
	bufio.Head = 88
	bufio.Tail = 8103
	bufio.OutHead = 8172
	bufio.DupSize = 0
	bufio.Checked = 8123

	if bufio.UnCheckedSeqLen() != 49 {
		t.Error("UnCeckedLen invalid  ")
	}

	bufio.Check(69)

	if bufio.OutHead == 0 || bufio.Checked != 20 {
		t.Error("cehck invalid  ", bufio.P())
	}

}
func TestOutRange(t *testing.T) {
	CreateFile(4000)

	bufio := NewTinyRBuff(4096*2, 20)
	bufio.ReadMax = 4096
	bufio.Head = 7144
	bufio.Tail = 6870
	bufio.OutHead = 0
	bufio.DupSize = 0
	bufio.Checked = 7027

	bufio.UnCheckedLen()
	bufio.Checkv(117)

}

func TestOverCheck(t *testing.T) {
	CreateFile(4000)

	bufio := NewTinyRBuff(4096*2, 20)
	bufio.ReadMax = 4096
	bufio.Head = 68
	bufio.Tail = 8103
	bufio.OutHead = 8172
	bufio.DupSize = 20
	bufio.Checked = 68

	if bufio.UnCheckedSeqLen() != 0 {
		t.Error("t.Checked overrun ", bufio.UnCheckedSeqLen())
	}

	bufio.Checked = 20

	if bufio.UnCheckedLen() != 48 {
		t.Error("t.Checked over UnCheckedLen ", bufio.UnCheckedLen(), bufio.P())
	}
	bufio.Check(48)

	if bufio.Checked != 68 {
		t.Error("t.Checked overrun ", bufio.P())
	}

}

func TestCheckUnderOverhead(t *testing.T) {
	CreateFile(4000)

	bufio := NewTinyRBuff(4096*2, 20)
	bufio.ReadMax = 4096
	bufio.Head = 0
	bufio.Tail = 7603
	bufio.OutHead = 8172
	bufio.DupSize = 0
	bufio.Checked = 8034

	old := bufio.UnCheckedLen()
	if bufio.Checkv(117); old != bufio.UnCheckedLen()+117 {
		t.Error("t.CheckUnderOverhead orver bufi.Checkv ", old, bufio.UnCheckedLen())
	}
}

func TestFill(t *testing.T) {
	CreateFile(4000)
	file, _ := os.Open("trbtest")
	outfile, _ := os.Create("outtest")

	bufio := NewTinyRBuff(4096*2, 20)
	bufio.ReadMax = 4096
	bufio.Head = 8103
	bufio.Tail = 8103
	bufio.OutHead = 0
	bufio.DupSize = 0

	_, e := bufio.ReadAtLeast(file, 20)
	fmt.Println("after read bufio", bufio.p())

	return

	if e != nil && bufio.UnCheckedSeqLen() < 1 {
		t.Error("ReadAtLeast", e, bufio.UnCheckedSeqLen())
	}
	bufio.Check(bufio.UnCheckedSeqLen())
	fmt.Println("after check bufio", bufio.p())

	written, _ := bufio.WriteAt(outfile, 0)
	fmt.Println("bufio written", bufio.p(), written)
	fmt.Println("UnCeckedLne()", bufio.UnCheckedLen())
	bufio.Check(bufio.UnCheckedLen())
	fmt.Println("recheck bufio ", bufio.p())
	outfile.Close()

	file, _ = os.Open("outtest")
	data := make([]byte, 30)
	file.Read(data[0:30])

	if data[20] != ([]byte)("1")[0] {
		t.Error("data[3] != ", ([]byte)("4"))
	}
	file.Close()
}
