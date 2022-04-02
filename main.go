package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/valyala/fasthttp"
)

func clearRest(writer io.Writer) {
	_, err := io.WriteString(writer, "\x1B[J")
	panicOnErr(err)
}

func panicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}

var alphaNum []uint8 = func() (out []uint8) {
	charRange := func(from uint8, to uint8) {
		for ch := from; ch <= to; ch++ {
			out = append(out, ch)
		}
	}
	charRange('a', 'z')
	charRange('A', 'Z')
	charRange('0', '9')
	return
}()

var foundOk uint32
var foundBad uint32

func fetch(
	config userConfig,
	urlChan chan<- string,
	errChan chan<- string,
) {
	gen := createGen(config)
	var (
		req    fasthttp.Request
		resp   fasthttp.Response
		url    fasthttp.URI
		client fasthttp.Client
		err    error
	)

	url.SetScheme("https")
	url.SetHost("i.imgur.com")
	client.Dial = func(addr string) (net.Conn, error) {
		return fasthttp.DialTimeout(addr, time.Minute)
	}

	for {
		path := gen.next()
		strPath := *(*string)(unsafe.Pointer(&path))
		url.SetPath(strPath)

		panicOnErr(err)
		req.SetURI(&url)
		req.Header.SetMethod(fasthttp.MethodHead)
		err = client.Do(&req, &resp)
		switch code := resp.StatusCode(); code {
		case 200:
			urlChan <- url.String()
			atomic.AddUint32(&foundOk, 1)
		case 409:
			str := strings.Builder{}
			_, _ = fmt.Fprintf(&str, "\rTimed out for sending too many requests")
			clearRest(&str)
			errChan <- str.String()
			return

		default:
			atomic.AddUint32(&foundBad, 1)
			if config.Logger != nil {
				config.Logger.Printf("Image not found at %v. Status Code: %d\n", url.String(), code)
			}
		}
	}
}

func main() {

	cfg := parseArgs()

	urlChan := make(chan string, 200)
	errChan := make(chan string)

	for i := 0; i < cfg.Connections; i++ {
		go fetch(cfg, urlChan, errChan)
	}

	out := bufio.NewWriter(os.Stdout)
	display := func(str string) {
		panicOnErr(out.WriteByte('\r'))
		_, err := out.WriteString(str)
		panicOnErr(err)
		clearRest(out)
	}

	prog := Progress{
		Interval: time.Second / 2,
		States: []string{
			"", ".", "..", "...",
		},
	}
	defer func() {
		_, _ = fmt.Fprintf(
			out, "OK: %d. Bad: %d. Percentage OK: %3.2f%%",
			foundOk, foundBad,
			float32(foundOk)/float32(foundOk+foundBad)*100,
		)
		_ = out.Flush()
	}()

	for recv := 0; recv < cfg.Amount; {
		select {
		case msg := <-errChan:
			_, err := fmt.Fprintf(os.Stderr, "%s\n", msg)
			panicOnErr(err)
			return

		case url := <-urlChan:
			recv++
			display(url)
			panicOnErr(out.WriteByte('\n'))

		default:
			str, same := prog.Next()
			if same {
				continue
			}
			display(str)
		}
		panicOnErr(out.Flush())
	}

}
