package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

func pushIt(command string, queue string, parametersString string, test bool, timeout int, verbose bool) {

	var split []string
	var filenames []string
	var placeholders []string

	parameters := strings.Split(parametersString, ",")
	for _, p := range parameters {
		split = strings.Split(p, ":")
		placeholders = append(placeholders, split[0])
		filenames = append(filenames, split[1])
	}

	var fileSlices [][]string

	for _, filename := range filenames {
		newFileLines, err := readLines(filename)
		if err != nil {
			log.Fatal(err)
		}
		fileSlices = append(fileSlices, newFileLines)
	}

	var lengths []int
	var originalLengths []int

	for i := range fileSlices {
		length := len(fileSlices[i])
		lengths = append(lengths, length)
		originalLengths = append(originalLengths, length)
	}

	var wg sync.WaitGroup

	queueID := uuid.New().String()

	// get the results back and print them
	go printResults(queueID, &wg, verbose)

	// this is the recursive function that generates all of the command combinations
	loopThrough(fileSlices, placeholders, command, lengths, originalLengths, test, &wg, queueID, timeout, queue)

	// wait for all results to be returned and printed before exiting
	wg.Wait()
}

func printResults(queueID string, wg *sync.WaitGroup, verbose bool) {
	for {
		result, err := redisClient.RPop(queueID).Result()
		if err != nil {
			if verbose {
				log.Println("Awaiting output.")
			}
			time.Sleep(1 * time.Second)
		} else {
			fmt.Println(result)
			wg.Done()
		}
	}
}

// checks if all elements of a slice are equal to comparator
func checkIfAll(array []int, comparator int) bool {
	for _, i := range array {
		if i != comparator {
			return false
		}
	}
	return true
}

// Recursive function that takes slices of strings, and prints every combination of lines in each file
func loopThrough(fileSlices [][]string, placeholders []string, command string, lengths []int, originalLengths []int, test bool, wg *sync.WaitGroup, queueID string, timeout int, queue string) {
	if checkIfAll(lengths, 0) {
		return
	}
	line := command

	for i, fileSlice := range fileSlices {

		// if the RHS number is 0, finish
		if lengths[len(lengths)-1] == 0 {
			return
		}

		// replace the placeholders in the line
		line = strings.ReplaceAll(line, "_"+placeholders[i]+"_", fileSlice[lengths[i]-1])
	}

	// if -test is specified, just print the commands, otherwise push them to redis
	wg.Add(1)
	if test {
		fmt.Println(line)
	} else {
		// using :::_::: as a separator between the queueID, timeout and command
		redisClient.LPush(queue, queueID+":::_:::"+strconv.Itoa(timeout)+":::_:::"+line)
	}

	for i := range lengths {
		// if our current number is a 0, decrement the number directly to the right, then set this back to the original
		if lengths[i] == 1 && len(lengths) != i+1 {
			lengths[i+1] = lengths[i+1] - 1
			lengths[i] = originalLengths[i]
			break
		} else {
			lengths[i] = lengths[i] - 1
			break
		}
	}

	loopThrough(fileSlices, placeholders, command, lengths, originalLengths, test, wg, queueID, timeout, queue)
}

// readLines reads a whole file into memory and returns a slice of its lines.
func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}
