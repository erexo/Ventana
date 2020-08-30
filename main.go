package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime/debug"

	"github.com/Erexo/Ventana/core/domain"
	"github.com/Erexo/Ventana/infrastructure/gpio"
)

func main() {
	fmt.Println("Hello")
	defer terminate()

	pm := gpio.CreatePinManager()

	c := make(chan os.Signal, 1)
	signal.Notify(c)
	go func() {
		x := <-c
		log.Println("Terminated", x)
		err := pm.Close()
		if err != nil {
			log.Println("PinManager Close failed:", err)
		}
		os.Exit(0)
	}()

	register(pm, 0, 2)
	register(pm, 1, 3)

	fmt.Fscanln(os.Stdout)
	fmt.Println("Bye")
}

func terminate() {
	if r := recover(); r != nil {
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

func register(pm *gpio.PinManager, input, output uint8) error {
	ip, err := domain.CreateMcpPin(0, input)
	if err != nil {
		return err
	}
	op, err := domain.CreateMcpPin(0, output)
	if err != nil {
		return err
	}
	return pm.AddPinPair(ip, op)
}
