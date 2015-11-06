package tiny_ring_bufio

import (
	"fmt"
	"os"
	"strconv"
	"testing"
)

func CreateFile(size int) {

	file, _ := os.Create("trbtest")
	for i := 0; i < size; i++ {
		file.Write(([]byte)(strconv.Itoa(i)))
	}
	file.Close()
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

//func Test
