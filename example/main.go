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

		return nil
	}

	sc := scheduler.NewScheduleJob("test_job", time.Second, fnJob)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	sc.Start(ctx, false)

	<-ctx.Done()
}
