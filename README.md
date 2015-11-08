# tiny_ring_bufio [![wercker status](https://app.wercker.com/status/b3b5360fed88e70a49bb4ccbc779adbf/s "wercker status")](https://app.wercker.com/project/bykey/b3b5360fed88e70a49bb4ccbc779adbf)

This pcakge provide Ring Buffer I/O wrap an io>Reader/io.Writer object and buffer ringed byte slice

sequence is 
read/check(parse)/write  .

# example
```go
file, _ := os.Open("test")
bufio := tiny_ring_bufio.New(1024, 30)
file_n, e := bufio.ReadAtLeast(file, 20) // read over 20byte

data := bufio.Check(bufio.UnCheckedSeqLen()) // get data as []byte

fmt.Println("dump bufio withouf buffer data", bufio.p())

```


# Contributing

Bug reports and pull requests are welcome on GitHub at  https://github.com/kazu/tiny_ring_bufio
