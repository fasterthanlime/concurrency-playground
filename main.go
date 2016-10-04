package main

import (
	"fmt"
	"log"
	"os"
	"time"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	app      = kingpin.New("concurrency-playground", "Playing around with goroutines and channels")
	doSelect = app.Flag("select", "Use a select statement when sending tasks").Default().Bool()
	fail     = app.Flag("fail", "Fail tasks that are a modulo of this number").Default("-1").Int()
)

func main() {
	_, err := app.Parse(os.Args[1:])
	if err != nil {
		panic(err)
	}

	done := make(chan bool)

	go func() {
		err = doStuff()
		if err != nil {
			log.Printf("got error %s", err.Error())
			os.Exit(1)
		}

		done <- true
	}()

	timeout := time.After(100 * time.Millisecond)

	select {
	case <-timeout:
		log.Printf("timed out!")
		os.Exit(1)
	case <-done:
		log.Printf("all done!")
	}
}

func doStuff() error {
	done := make(chan bool)
	errs := make(chan error)

	tasks := make(chan int)
	numWorkers := 8

	for i := 0; i < numWorkers; i++ {
		go doWork(tasks, done, errs)
	}

	for i := 1; i < 32; i++ {
		if *doSelect {
			select {
			case err := <-errs:
				close(tasks)
				return err
			case tasks <- i:
				// good!
			}
		} else {
			tasks <- i
		}
	}

	close(tasks)

	for i := 0; i < numWorkers; i++ {
		select {
		case err := <-errs:
			return err
		case <-done:
			// good!
		}
	}

	return nil
}

func doWork(tasks chan int, done chan bool, errs chan error) {
	for task := range tasks {
		log.Printf("working %d", task)
		time.Sleep(time.Duration(task) * time.Millisecond)

		if (*fail) != -1 {
			if task%(*fail) == 0 {
				errs <- fmt.Errorf("failed %d", task)
			}
		}
		log.Printf("done working %d", task)
	}

	done <- true
}
