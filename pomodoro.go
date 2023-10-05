package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

func main() {

	args := os.Args[1:]
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "No task ID passed")
		os.Exit(1)
	}

	taskWarriorID := args[0]
	var intervalInMins int
	if len(args) < 2 {
		interval, err := strconv.Atoi(textLinePrompt("For how many minutes?"))
		intervalInMins = interval
		if err != nil {
			fmt.Fprintln(os.Stderr, "Invalid time interval: ", err)
			os.Exit(2)
		}
	} else {
		interval, err := strconv.Atoi(args[1])
		intervalInMins = interval
		if err != nil {
			fmt.Fprintln(os.Stderr, "Invalid time interval: ", err)
			os.Exit(2)
		}
	}

	b, err := exec.Command("task", "_get", fmt.Sprintf("%s.description", taskWarriorID)).Output()
	if err != nil {
		fmt.Println("task warrior error: ", err)
		return
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go handleInterrupt(sigs, taskWarriorID)

	err = exec.Command("task", "start", taskWarriorID).Run()
	if err != nil {
		fmt.Println("error starting task: ", err)
		done(taskWarriorID)
		return
	}

	fmt.Printf("Starting '%s' for %d minutes\n", string(b), intervalInMins)
	waitDuration := time.Duration(intervalInMins) * time.Minute
	fmt.Println("Stopping at: ", time.Now().Add(waitDuration).Format(time.Kitchen))
	time.Sleep(waitDuration)

	fmt.Println("Pomodoro completed.")

	wg := sync.WaitGroup{}
	wg.Add(1)
	go playSound(&wg)
	done(taskWarriorID)
	wg.Wait()
}

func playSound(wg *sync.WaitGroup) {
	for i := 0; i < 5; i++ {
		exec.Command("afplay", "/System/Library/Sounds/Submarine.aiff").Run()
	}
	wg.Done()
}

func handleInterrupt(sigs chan os.Signal, taskID string) {
	sig := <-sigs
	fmt.Println()
	fmt.Println("Received interrupt: ", sig)
	done(taskID)
	os.Exit(1)
}

func done(taskID string) {
	fmt.Println("Stopping task")
	exec.Command("task", taskID, "stop").Run()
}

func textLinePrompt(msg string) string {
	print(fmt.Sprintf("%v: ", msg))
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	input := scanner.Text()
	return input
}

func confirmPrompt(msg string) (bool, error) {
	fmt.Println(msg)
	var input string
	print("y/N: ")
	_, e := fmt.Scanf("%s", &input)
	if e != nil {
		return false, e
	}
	normalisedInput := strings.ToLower(strings.TrimSpace(input))
	if normalisedInput == "y" || normalisedInput == "yes" {
		return true, e
	}
	return false, e
}
