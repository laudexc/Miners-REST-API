package logic

import (
	"context"
	"fmt"
	"sync"
	"time"
)

func basedEnterprise(ctx context.Context, wg *sync.WaitGroup, transferPoint chan<- int) {
	defer wg.Done()

	for {
		fmt.Println("Базовое предприятие начинает работу и теперь приносит +1 уголь каждую секунду")

		select {
		case <-ctx.Done():
			// TODO: почему предприятие может закончить работу?
			fmt.Println("Базовое предприятие закончило работу по какой то причине")
			return

		case <-time.After(1 * time.Second):
			// NOTE: Заглушка, не знаю что тут написать чтобы не было спама
			fmt.Print()
		}

		select {
		case <-ctx.Done():
			// TODO: почему предприятие может закончить работу?
			fmt.Println("Базовое предприятие закончило работу по какой то причине")

		case transferPoint <- 1:
			fmt.Println("Базовое предприятие принесло 1 уголь")
		}
	}
}

func basedEntpPool(ctx context.Context) <-chan int {
	coalTransferPoint := make(chan int)
	wg := &sync.WaitGroup{}

	wg.Add(1)
	go basedEnterprise(ctx, wg, coalTransferPoint)

	// TODO: При каком условии это работает?
	go func() {
		wg.Wait()
		close(coalTransferPoint)
	}()

	return coalTransferPoint
}
