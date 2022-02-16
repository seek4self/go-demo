// loop queue
package main

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

var q = newQueue(3)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	ch := make(chan struct{}, 3)
	wg := &sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		ch <- struct{}{}
		wg.Add(1)
		go do(i, ch, wg)
	}
	wg.Wait()
	fmt.Println(q.gpu)

}

func do(i int, ch chan struct{}, wg *sync.WaitGroup) {
	n := q.getFront()
	log.Println(n, i)
	time.Sleep(time.Duration(rand.Intn(7)+3) * time.Second)
	q.append(n)
	wg.Done()
	<-ch
}

type queue struct {
	m           *sync.Mutex
	front, tail int
	len         int
	gpu         []int
}

func newQueue(n int) *queue {
	q := &queue{
		m:   &sync.Mutex{},
		len: n,
		gpu: make([]int, 0, n),
	}
	for i := 0; i < n; i++ {
		q.gpu = append(q.gpu, i)
		q.tail = (q.tail + 1) % q.len
	}
	return q
}

func (q *queue) getFront() int {
	q.m.Lock()
	defer q.m.Unlock()
	n := q.gpu[q.front]
	q.gpu[q.front] = -1
	log.Println(q.gpu, q.front, q.tail)
	q.front = (q.front + 1) % q.len
	return n
}

func (q *queue) append(n int) {
	q.m.Lock()
	defer q.m.Unlock()
	q.gpu[q.tail] = n
	q.tail = (q.tail + 1) % q.len
	log.Println("set", q.tail, "to", n)
}
