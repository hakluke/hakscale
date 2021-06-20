package main

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

func popIt(threads int, queue string, verbose bool) {

	var wg sync.WaitGroup

	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func() {
			doWork(&wg, queue, verbose)
		}()
	}

	wg.Wait()
}

func doWork(wg *sync.WaitGroup, queue string, verbose bool) {
	defer wg.Done()
	for {
		result, err := redisClient.RPop(queue).Result()
		if err != nil {
			if verbose {
				log.Println("Polling for jobs.")
			}
			time.Sleep(1 * time.Second)
		} else {
			shellexec(result, verbose)
		}
	}
}

func shellexec(command string, verbose bool) {
	ctx := context.Background()
	split := strings.Split(command, ":::_:::")
	queue := split[0] // this is a randomly generated uuid to be used as a queue name for returning the output
	timeout, err := strconv.Atoi(split[1])
	if err != nil {
		log.Println(err)
		redisClient.LPush(queue, err.Error) // push error to the queue
		return
	}
	command = split[2]

	if verbose {
		log.Println("Running command:", command)
	}

	commandctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(timeout)) // add context to include a timeout
	defer cancel()                                                                                      // The cancel should be deferred so resources are cleaned up
	out, err := exec.CommandContext(commandctx, "/bin/bash", "-c", command).Output()
	if commandctx.Err() == context.DeadlineExceeded {
		writeToQueueAndPrint(ctx, command, queue, []byte("Command timed out.\n")) // push error to the queue
	} else if err != nil {
		writeToQueueAndPrint(ctx, command, queue, []byte(err.Error())) // push error to the queue
	} else {
		writeToQueueAndPrint(ctx, command, queue, out)
	}
}

// writeToQueueAndPrint will print the command output and then write it to the redis queue
func writeToQueueAndPrint(ctx context.Context, command string, queue string, output []byte) {
	log.Println("Output for command:", command)
	fmt.Println(string(output))      // print command output
	redisClient.LPush(queue, output) // push the command output to the queue
}
