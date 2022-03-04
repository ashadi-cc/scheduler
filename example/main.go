package main

import (
	"context"
	"log"
	"time"

	"github.com/ashadi-cc/scheduler"
)

func main() {
	fnJob := func(ctx context.Context) error {
		log.Println("Run function")

		panic("wat panic")
	}

	sc := scheduler.NewScheduleJob("test_job", time.Second, fnJob)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	sc.Start(ctx, 2)

	<-ctx.Done()
}
