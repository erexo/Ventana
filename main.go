package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime/debug"

	"github.com/Erexo/Ventana/api"
	"github.com/Erexo/Ventana/infrastructure/gpio"
	"github.com/Erexo/Ventana/infrastructure/sunblind"
)

func main() {
	fmt.Println("Hello")
	defer panic()

	pm := gpio.CreatePinManager()

	c := make(chan os.Signal, 1)
	signal.Notify(c)
	go terminate(pm, c)

	ss := sunblind.CreateService(pm)
	if err := ss.Load(); err != nil {
		log.Println("SunblindService error:", err)
	}

	// todo, add flag to run api
	if err := api.Run(ss); err != nil {
		log.Println("Api error:", err)
	}

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

func terminate(pm *gpio.PinManager, c chan os.Signal) {
	x := <-c
	log.Println("Terminated", x)
	err := pm.Close()
	if err != nil {
		log.Println("PinManager Close failed:", err)
	}
	os.Exit(0)
}
