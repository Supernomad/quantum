package main

import (
	"github.com/Supernomad/quantum/agg"
	"github.com/Supernomad/quantum/backend"
	"github.com/Supernomad/quantum/common"
	"github.com/Supernomad/quantum/inet"
	"github.com/Supernomad/quantum/socket"
	"github.com/Supernomad/quantum/workers"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
)

const version string = "0.6.0"

func handleError(log *common.Logger, err error, stack string) {
	if err != nil {
		log.Error.Println(err.Error(), "Stack:", stack)
		os.Exit(1)
	}
}

func main() {
	log := common.NewLogger()
	wg := &sync.WaitGroup{}

	cfg, err := common.NewConfig()
	handleError(log, err, "common.NewConfig()")

	store, err := backend.New(backend.LIBKV, log, cfg)
	handleError(log, err, "backend.New(backend.LIBKV, log, cfg)")

	err = store.Init()
	handleError(log, err, "store.Init()")

	tunnel := inet.New(inet.TUNInterface, cfg)
	err = tunnel.Open()
	handleError(log, err, "tunnel.Open()")

	sock := socket.New(socket.UDPSocket, cfg)
	err = sock.Open()
	handleError(log, err, "sock.Open()")

	outgoing := workers.NewOutgoing(cfg, store, tunnel, sock)

	incoming := workers.NewIncoming(cfg, store, tunnel, sock)

	aggregator := agg.New(log, cfg, incoming.QueueStats, outgoing.QueueStats)

	wg.Add(2)
	aggregator.Start(wg)
	store.Start(wg)
	for i := 0; i < cfg.NumWorkers; i++ {
		incoming.Start(i)
		outgoing.Start(i)
	}

	log.Info.Println("[MAIN]", "Listening on TUN device:  ", tunnel.Name())
	log.Info.Println("[MAIN]", "TUN network space:        ", cfg.NetworkConfig.Network)
	log.Info.Println("[MAIN]", "TUN private IP address:   ", cfg.PrivateIP)
	log.Info.Println("[MAIN]", "TUN public IPv4 address:    ", cfg.PublicIPv4)
	log.Info.Println("[MAIN]", "TUN public IPv6 address:    ", cfg.PublicIPv6)
	log.Info.Println("[MAIN]", "Listening on UDP port: ", strconv.Itoa(cfg.ListenPort))

	exit := make(chan struct{})
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	go func() {
		sig := <-signals
		switch {
		case sig == syscall.SIGHUP:
			log.Info.Println("[MAIN]", "Recieved reload signal from user. Reloading process.")

			sockFDS := sock.GetFDs()
			tunFDS := tunnel.GetFDs()

			files := make([]uintptr, 3+cfg.NumWorkers*2)
			files[0] = os.Stdin.Fd()
			files[1] = os.Stdout.Fd()
			files[2] = os.Stderr.Fd()

			for i := 0; i < cfg.NumWorkers; i++ {
				files[3+i] = uintptr(tunFDS[i])
				files[3+i+cfg.NumWorkers] = uintptr(sockFDS[i])
			}

			os.Setenv(common.RealInterfaceNameEnv, tunnel.Name())
			pid, err := syscall.ForkExec(os.Args[0], os.Args, &syscall.ProcAttr{
				Env:   os.Environ(),
				Files: files,
			})
			handleError(log, err, "syscall.ForkExec(os.Args[0], os.Args, &syscall.ProcAttr{Env: os.Environ(), Files: files})")

			ioutil.WriteFile(cfg.PidFile, []byte(strconv.Itoa(pid)), 0644)
		case sig == syscall.SIGINT || sig == syscall.SIGTERM || sig == syscall.SIGKILL:
			log.Info.Println("[MAIN]", "Recieved termination signal from user. Terminating process.")
		}
		exit <- struct{}{}
	}()
	<-exit

	aggregator.Stop()
	store.Stop()
	incoming.Stop()
	outgoing.Stop()

	sock.Close()
	tunnel.Close()
	wg.Wait()
}
