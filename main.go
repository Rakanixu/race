package main

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type goroutine struct {
	id       int
	progress int
}

type goroutines []goroutine

var monitor goroutines

func main() {
	var wg sync.WaitGroup
	c := make(chan int)
	d := make(chan goroutine)

	w, h := stdoutSize()
	clear(w, h)

	go drawState(d)

	for i := 1; i < h; i++ {
		wg.Add(1)
		monitor = append(monitor, goroutine{
			id:       i,
			progress: 0,
		})

		go draw(i, 0, w, c, d)
	}

	go func() {
		for {
			select {
			case <-c:
				wg.Done()
			}
		}
	}()

	wg.Wait()

	fmt.Printf("\x1b[39;49m")
	fmt.Printf("\033[%d;%dH ", w, h)
}

func draw(y, x, limit int, done chan int, data chan goroutine) {
	char := "="
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	fmt.Printf("\x1b[27;33m")
	for {
		if len(char) >= limit-5 {
			done <- 1
			return
		}
		fmt.Printf("\033[%d;%dH %d.%s>", y+1, x, y, char)
		char = strings.Join([]string{char, "="}, "")
		data <- goroutine{
			id:       y,
			progress: len(char),
		}
		time.Sleep(time.Duration(r.Int31() / 8))
	}
}

func drawState(c chan goroutine) {
	for {
		select {
		case val := <-c:
			monitor[val.id-1] = val
			sort.Sort(monitor)
			fmt.Printf("\033[0;0H Leading: %d                                   ", monitor[0].id)
		}
	}
}

func clear(w, h int) {
	var buffer bytes.Buffer

	for i := 0; i < w; i++ {
		buffer.Write([]byte(" "))
	}

	fmt.Printf("\033[0;0H")
	for i := 0; i < h; i++ {
		fmt.Printf(buffer.String())
	}
}

func stdoutSize() (int, int) {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	handleErr(err)

	s := strings.Split(string(out)[:len(string(out))-1], " ")
	h, err := strconv.Atoi(s[0])
	handleErr(err)
	w, err := strconv.Atoi(s[1])
	handleErr(err)

	return w, h
}

func handleErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func (g goroutines) Len() int {
	return len(g)
}

func (g goroutines) Less(i, j int) bool {
	return g[i].progress > g[j].progress
}

func (g goroutines) Swap(i, j int) {
	g[i], g[j] = g[j], g[i]
}
