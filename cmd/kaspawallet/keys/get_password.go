// Adapted from https://gist.github.com/jlinoff/e8e26b4ffa38d379c7f1891fd174a6d0#file-getpassword2-go
package keys

import (
	"fmt"
	"golang.org/x/term"
	"os"
	"os/signal"
	"syscall"
)

func getPassword(prompt string) string {
	// Get the initial state of the terminal.
	initialTermState, e1 := term.GetState(syscall.Stdin)
	if e1 != nil {
		panic(e1)
	}

	// Restore it in the event of an interrupt.
	// CITATION: Konstantin Shaposhnikov - https://groups.google.com/forum/#!topic/golang-nuts/kTVAbtee9UA
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill)
	go func() {
		<-c
		_ = term.Restore(syscall.Stdin, initialTermState)
		os.Exit(1)
	}()

	// Now get the password.
	fmt.Print(prompt)
	p, err := term.ReadPassword(syscall.Stdin)
	fmt.Println("")
	if err != nil {
		panic(err)
	}

	// Stop looking for ^C on the channel.
	signal.Stop(c)

	// Return the password as a string.
	return string(p)
}
