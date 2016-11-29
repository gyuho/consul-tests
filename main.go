package main

import (
	"flag"
	"fmt"
	mrand "math/rand"
	"strings"
	"sync"
	"time"

	consulapi "github.com/hashicorp/consul/api"
)

func main() {
	endpointsPt := flag.String("endpoints", "localhost:8500", "endpoints (currently supports only 1 endpoint)")
	clientsNPt := flag.Int("clients", 100, "number of concurrent clients")
	writesNPt := flag.Int("writes", 1000, "number of write requests")
	keySizePt := flag.Int("key-size", 8, "key size")
	valSizePt := flag.Int("val-size", 256, "value size")
	flag.Parse()

	eps := []string{*endpointsPt}
	clientsN := *clientsNPt
	writesN := *writesNPt
	keySize := *keySizePt
	valSize := *valSizePt

	var wg sync.WaitGroup

	pairs := make(chan consulapi.KVPair, clientsN) // buffer == maximum concurrent clients size
	wg.Add(1)
	go func() {
		defer func() {
			wg.Done()
			close(pairs)
		}()

		val := randBytes(valSize)
		for i := 0; i < writesN; i++ {
			pairs <- consulapi.KVPair{Key: sequentialKey(keySize, i), Value: val}
		}
	}()

	results := make(chan result, writesN)
	donec, errm := make(chan struct{}), make(map[string]int)
	total, min, max := time.Duration(0), time.Duration(0), time.Duration(0)
	cnt := 0
	go func() {
		defer close(donec)
		for rs := range results {
			if cnt%500 == 0 {
				fmt.Println("success", cnt, "/", writesN)
			}
			if rs.errStr != "" {
				errm[rs.errStr]++
			}
			total += rs.duration
			if min == time.Duration(0) {
				min = rs.duration
			}
			if min > rs.duration {
				min = rs.duration
			}
			if max < rs.duration {
				max = rs.duration
			}
			cnt++
		}
	}()

	conns := mustCreateConnsConsul(eps, clientsN)
	for i := range conns {
		wg.Add(1)
		go func(i int, conn *consulapi.KV) {
			defer wg.Done()
			for pair := range pairs {
				reqNow := time.Now()
				_, err := conn.Put(&pair, nil)
				reqTook := time.Since(reqNow)
				errStr := ""
				if err != nil {
					errStr = err.Error()
				}
				results <- result{errStr: errStr, duration: reqTook}
			}
		}(i, conns[i])
	}
	wg.Wait()
	close(results)
	<-donec

	fmt.Println("Total clients:", clientsN)
	fmt.Println("Total requests:", writesN)
	fmt.Println("Average:", total/time.Duration(writesN))
	fmt.Println("Min:", min)
	fmt.Println("Max:", max)

	if len(errm) == 0 {
		fmt.Println("No error!")
	} else {
		fmt.Println("Error:")
		for k, v := range errm {
			fmt.Println(k, v)
		}
	}

	// for i := 0; i < 10; i++ {
	// 	resp, _, err := conns[0].Get(sequentialKey(keySize, i), nil)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	rs := *resp
	// 	fmt.Printf("Get response: %q %q\n", rs.Key, string(rs.Value))
	// }
}

var dialTotal int

func mustCreateConnsConsul(endpoints []string, total int) []*consulapi.KV {
	css := make([]*consulapi.KV, total)
	for i := range css {
		endpoint := endpoints[dialTotal%len(endpoints)]
		dialTotal++

		dcfg := consulapi.DefaultConfig()
		dcfg.Address = endpoint // x.x.x.x:8500
		cli, err := consulapi.NewClient(dcfg)
		if err != nil {
			panic(err)
		}

		css[i] = cli.KV()
	}
	return css
}

// sequentialKey returns '00012' when size is 5 and num is 12.
func sequentialKey(size, num int) string {
	txt := fmt.Sprintf("%d", num)
	if len(txt) > size {
		return txt
	}
	delta := size - len(txt)
	return strings.Repeat("0", delta) + txt
}

func randBytes(bytesN int) []byte {
	const (
		letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
		letterIdxBits = 6                    // 6 bits to represent a letter index
		letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
		letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
	)
	src := mrand.NewSource(time.Now().UnixNano())
	b := make([]byte, bytesN)
	for i, cache, remain := bytesN-1, src.Int63(), letterIdxMax; i >= 0; {
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
	return b
}

type result struct {
	errStr   string
	duration time.Duration
}
