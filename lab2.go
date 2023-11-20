package main

import (
	"context"
	"fmt"
	"golang.org/x/sync/semaphore"
	"sync"
)

func parallelSearch(arr []int, target int, workers int) int {
	var resultIdx int
	var wg sync.WaitGroup
	var mu sync.Mutex
	ch := make(chan int, workers)

	worker := func(startIdx, endIdx int) {
		defer wg.Done()
		for i := startIdx; i < endIdx; i++ {
			if arr[i] == target {
				ch <- i
				return
			}
		}
	}

	chunkSize := len(arr) / workers

	wg.Add(workers)
	for i := 0; i < workers; i++ {
		startIdx := i * chunkSize
		endIdx := (i + 1) * chunkSize
		if i == workers-1 {
			endIdx = len(arr)
		}
		go worker(startIdx, endIdx)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for idx := range ch {
		mu.Lock()
		resultIdx = idx
		mu.Unlock()
		break
	}

	return resultIdx
}

func threadPoolSearch(arr []int, target int, maxWorkers int) int {
	var resultIdx int
	var wg sync.WaitGroup
	var mu sync.Mutex
	sem := semaphore.NewWeighted(int64(maxWorkers))

	worker := func(startIdx, endIdx int) {
		defer wg.Done()
		defer sem.Release(1)

		for i := startIdx; i < endIdx; i++ {
			if arr[i] == target {
				mu.Lock()
				resultIdx = i
				mu.Unlock()
				return
			}
		}
	}

	chunkSize := len(arr) / maxWorkers

	for i := 0; i < maxWorkers; i++ {
		startIdx := i * chunkSize
		endIdx := (i + 1) * chunkSize
		if i == maxWorkers-1 {
			endIdx = len(arr)
		}

		if err := sem.Acquire(context.TODO(), 1); err != nil {
			fmt.Println("Не удалось приобрести семафор:", err)
			return -1
		}

		wg.Add(1)
		go worker(startIdx, endIdx)
	}

	wg.Wait()

	return resultIdx
}

func main() {
	arr := []int{16, 22, 43, 44, 15, 676, 75, 89, 933, 210}
	target := 89
	workers := 4
	maxWorkers := 2

	fmt.Println("Поиск с использованием Fork-Join Pool:")
	resultFJP := parallelSearch(arr, target, workers)
	if resultFJP != -1 {
		fmt.Printf("Элемент %d найден по индексу %d\n", target, resultFJP)
	} else {
		fmt.Printf("Элемент %d не найден\n", target)
	}

	fmt.Println("\nПоиск с использованием произвольного пула потоков:")
	resultTP := threadPoolSearch(arr, target, maxWorkers)
	if resultTP != -1 {
		fmt.Printf("Элемент %d найден по индексу %d\n", target, resultTP)
	} else {
		fmt.Printf("Элемент %d не найден\n", target)
	}
}
