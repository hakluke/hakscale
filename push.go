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

	// split the parameters input into a slice of placeholder:filename pairs
	parameters := strings.Split(parametersString, ",")

	// separate the placeholder from the filename and store them in two slices for processing
	for _, p := range parameters {
		split = strings.Split(p, ":")
		placeholders = append(placeholders, split[0])
		filenames = append(filenames, split[1])
	}

	// get a slice of lines for each filename
	var fileSlices [][]string
	for _, filename := range filenames {
		newFileLines, err := readLines(filename)
		if err != nil {
			log.Fatal(err)
		}
		fileSlices = append(fileSlices, newFileLines)
	}

	// for each file, figure out how many lines there are. Store lengths of each file in a slice called lengths, and do the same for originalLengths
	var lengths []int
	var originalLengths []int

	for i := range fileSlices {
		length := len(fileSlices[i])
		lengths = append(lengths, length)
		originalLengths = append(originalLengths, length)
	}

	var wg sync.WaitGroup

	// create a rendom queue name for sending back data
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

		switch {
		case err == redisClient.Nil: // the queue doesn't exist, there's no output to print yet
			if verbose {
				log.Println("Awaiting output:", err)
			}
			time.Sleep(1 * time.Second)
		case err != nil: // there was an actual error trying to grab the data
			log.Println("Redis error:", err)
		case result == "": // the command returned no output, don't bother printing a blank line
			wg.Done()
		default: // we got output, print it!
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
	if test {
		fmt.Println(line)
	} else {
		// using :::_::: as a separator between the queueID, timeout and command
		redisClient.LPush(queue, queueID+":::_:::"+strconv.Itoa(timeout)+":::_:::"+line)
		wg.Add(1)
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
