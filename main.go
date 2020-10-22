package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/Erexo/Ventana/infrastructure"
	"github.com/Erexo/Ventana/infrastructure/db"
)

func main() {
	fmt.Println("Hello")
	if runtime.GOOS == "windows" {
		fmt.Println("This application may not be executed on windows system")
		os.Exit(1)
	}
	if !strings.HasPrefix(runtime.GOARCH, "arm") {
		fmt.Println("This application may only be executed on arm architecture")
		os.Exit(1)
	}
	defer panic()

	err := db.Initialize()
	if err != nil {
		log.Fatal(err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c)
	go terminate(c)

	infrastructure.Run()

	fmt.Fscanln(os.Stdout)
	fmt.Println("Bye")
}

func panic() {
	if r := recover(); r != nil {
		// todo, check if panic do terminate
		var ok bool
		err, ok := r.(error)
		if !ok {
			err = fmt.Errorf("terminate: %v", r)
		}
		log.Println("Panic:", err)
		stack := string(debug.Stack())
		fmt.Println(stack)
	}
}

func terminate(c chan os.Signal) {
	x := <-c
	log.Println("Terminated", x)
	if err := infrastructure.Terminate(); err != nil {
		log.Println("Close failed:", err)
	}
	os.Exit(0)
}
