# tiny_ring_bufio [![wercker status](https://app.wercker.com/status/b3b5360fed88e70a49bb4ccbc779adbf/m "wercker status")](https://app.wercker.com/project/bykey/b3b5360fed88e70a49bb4ccbc779adbf)

tiny_ring_bufio implement ringed buffer I/O , wrap an io.Reader or  io.Writeer object  and buffer ringed byte slice


# example
```go
file, _ := os.Open("test")
bufio := tiny_ring_bufio.New(1024, 30)
file_n, e := bufio.ReadAtLeast(file, 20) // read over 20byte

data := bufio.Check(bufio.UnCheckedSeqLen()) // get data as []byte

fmt.Println("dump bufio withouf buffer data", bufio.p())

```go
