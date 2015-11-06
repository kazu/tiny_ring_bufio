# tiny_ring_bufio  [![wercker status](https://app.wercker.com/status/563cc3b46709dc217b0113f4/s "wercker status")] (https://app.wercker.com/project/bykey/563cc3b46709dc217b0113f4)

tiny_ring_bufio implement ringed buffer I/O , wrap an io.Reader or  io.Writeer object  and buffer ringed byte slice


'''
file, _ := os.Open("test")
bufio := tiny_ring_bufio.New(1024, 30)
file_n, e := bufio.ReadAtLeast(file, 20) // read over 20byte

data := bufio.Check(bufio.UnCheckedSeqLen()) // get data as []byte

fmt.Println("dump bufio withouf buffer data", bufio.p())
'''
