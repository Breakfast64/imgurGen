package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"
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

const imgurUrl = "https://i.imgur.com"

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
	client *http.Client,
	urlChan chan<- string,
	errChan chan<- string,
) {
	gen := createGen(config)
	req, err := http.NewRequest("HEAD", imgurUrl, nil)
	panicOnErr(err)

	for {
		path := gen.next()
		req.URL.Path = path
		resp, err := client.Do(req)
		panicOnErr(err)
		_ = resp.Body.Close()
		switch resp.StatusCode {
		case 200:
			urlChan <- resp.Request.URL.String()
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
				config.Logger.Printf("Image not found at %v. Status Code: %d\n", resp.Request.URL, resp.StatusCode)
			}
		}
	}
}

func main() {

	cfg := parseArgs()

	urlChan := make(chan string, 200)
	errChan := make(chan string)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	for i := 0; i < cfg.Connections; i++ {
		go fetch(cfg, client, urlChan, errChan)
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
