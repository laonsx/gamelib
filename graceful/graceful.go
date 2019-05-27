// +build !windows

package graceful

import (
	"context"
	"errors"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

var (
	server   *http.Server
	listener net.Listener
	graceful bool
)

func init() {

	flag.BoolVar(&graceful, "graceful", false, "graceful")
}

func reload(listener net.Listener) error {

	tl, ok := listener.(*net.TCPListener)
	if !ok {

		return errors.New("listener is not tcp listener")
	}

	f, err := tl.File()
	if err != nil {

		return err
	}

	args := os.Args[1:]
	if !contains(args, "--graceful=true") {

		args = append(args, "--graceful=true")
	}

	cmd := exec.Command(os.Args[0], args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.ExtraFiles = []*os.File{f}

	return cmd.Start()
}

func contains(s []string, e string) bool {

	for _, a := range s {

		if a == e {

			return true
		}
	}

	return false
}

func signalHandler(server *http.Server, listener net.Listener) {

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR2)
	for {

		sig := <-ch
		log.Println("[info] Signal received", "pid[", os.Getpid(), "] signal[", sig, "]")

		ctx, _ := context.WithTimeout(context.Background(), 20*time.Second)
		switch sig {

		case syscall.SIGINT, syscall.SIGTERM:

			log.Println("[warn] SignalHandler stop", "pid[", os.Getpid(), "] signal[", sig, "]")
			signal.Stop(ch)
			_ = server.Shutdown(ctx)

			log.Println("[warn] SignalHandler graceful shutdown", "pid[", os.Getpid(), "] signal[", sig, "]")

			return
		case syscall.SIGUSR2:
			log.Println("[warn] SignalHandler reload", "pid[", os.Getpid(), "] signal[", sig, "]")

			err := reload(listener)
			if err != nil {

				log.Println("[error] SignalHandler graceful failed", "pid[", os.Getpid(), "] signal[", sig, "] error[", err, "]")
			}
			_ = server.Shutdown(ctx)

			log.Println("[warn] SignalHandler graceful reload", "pid[", os.Getpid(), "] signal[", sig, "]")
			return
		}
	}
}

func ListenAndServe(addr string, handler http.Handler) {

	server = &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	var err error
	if graceful {

		log.Println("[info] Listening to existing file descriptor 3", "pid[", os.Getpid(), "] listen[", addr, "]")
		f := os.NewFile(3, "")
		listener, err = net.FileListener(f)
	} else {

		log.Println("[info] Listening on a new file descriptor", "pid[", os.Getpid(), "] listen[", addr, "]")
		listener, err = net.Listen("tcp", server.Addr)
	}

	if err != nil {

		log.Println("[error] listener error", "pid[", os.Getpid(), "] listen[", addr, "] error[", err, "]")

		return
	}

	go func() {
		err = server.Serve(listener)

		if err != nil {

			log.Println("[warn] server.Serve status", "pid[", os.Getpid(), "] listen[", addr, "] error[", err, "]")
		}
	}()

	signalHandler(server, listener)
}
