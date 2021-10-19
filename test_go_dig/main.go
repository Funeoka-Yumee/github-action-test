package main

import (
	"fmt"
	"net"
	"sync"
	"time"
	"unsafe"
	"math/rand"

	"github.com/miekg/dns"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
    letterIdxBits = 6                    // 6 bits to represent a letter index
    letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
    letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)
var src = rand.NewSource(time.Now().UnixNano())

func RandStringBytesMaskImprSrcUnsafe(n int) string {
	suffix := [...]byte{'.', 'a', '6', '0', '0', '8', '.', 'c','o','m','.'}
	length := len(suffix)
    b := make([]byte, n+length)
    // A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
    for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
        if remain == 0 {
            cache, remain = src.Int63(), letterIdxMax
        }
        if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
            b[i] = letterBytes[idx]
            i--
        }
        cache >>= letterIdxBits
        remain--
    }
	for i := 0; i < length; i++ {
		b[n+i] = suffix[i]
	}

    return *(*string)(unsafe.Pointer(&b))
}

func getMsg() (*dns.Msg) {
	m := &dns.Msg{}
	m.Id = dns.Id()
	m.RecursionDesired = true
	m.Question = make([]dns.Question, 1)
	for i := 0; i < len(m.Question); i++ {
		m.Question[i] = dns.Question{RandStringBytesMaskImprSrcUnsafe(5), dns.TypeA, dns.ClassINET}
	}

	return m
}

func queryOnce(msg *dns.Msg, server string) {
	raddr, err := net.ResolveUDPAddr("udp", server)
	if err != nil {
		fmt.Printf("failed to resolve udp addr %v\n", err)
		return
	}

	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		fmt.Printf("UDP connect error: %v\n", err)
		return
	}
	defer conn.Close()

	buf, err := msg.Pack()
	if err != nil {
		fmt.Printf("dns msg pack error %v\n", err)
		return
	}

	_, err = conn.Write(buf)
	if err != nil {
		fmt.Printf("failed to write query %v\n", err)
		return
	}

	ans := make([]byte, 4096)
	conn.SetReadDeadline(time.Now().Add(time.Second * 3))
	n, err := conn.Read(ans)
	if err != nil {
		fmt.Printf("udp read error: %v\n", err)
		return
	}
	m := &dns.Msg{}
	err = m.Unpack(ans[:n])
	if err != nil {
		fmt.Printf("dns parse error %v\n", err)
		return
	}
	if len(m.Answer) > 0 {
		fmt.Printf("got %d answer(s): %v...\n", len(m.Answer), m.Answer[0].String())
	} else {
		fmt.Printf("got empty answer\n")
	}
}

func worker(id int, jobs <-chan string, wg *sync.WaitGroup) {
	msg := getMsg()
	for server := range jobs {
		queryOnce(msg, server)
	}
	fmt.Printf("worker %d exits\n", id)
	wg.Done()
}

func main() {
	const workers = 16
	const server = "120.79.177.7:53"

	jobs := make(chan string, 256)
	wg := &sync.WaitGroup{}
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go worker(i, jobs, wg)
	}

	for i := 0; i < 4096; i++ {
		jobs <- server
	}
	close(jobs)

	wg.Wait()
}


