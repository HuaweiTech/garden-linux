package main

import (
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/cloudfoundry-incubator/garden/server"
	"github.com/cloudfoundry-incubator/warden-linux/linux_backend"
	"github.com/cloudfoundry-incubator/warden-linux/linux_backend/container_pool"
	"github.com/cloudfoundry-incubator/warden-linux/linux_backend/network_pool"
	"github.com/cloudfoundry-incubator/warden-linux/linux_backend/port_pool"
	"github.com/cloudfoundry-incubator/warden-linux/linux_backend/quota_manager"
	"github.com/cloudfoundry-incubator/warden-linux/linux_backend/uid_pool"
	"github.com/cloudfoundry/gunk/command_runner/linux_command_runner"
)

var listenNetwork = flag.String(
	"listenNetwork",
	"unix",
	"how to listen on the address (unix, tcp, etc.)",
)

var listenAddr = flag.String(
	"listenAddr",
	"/tmp/warden.sock",
	"address to listen on",
)

var snapshotsPath = flag.String(
	"snapshots",
	"",
	"directory in which to store container state to persist through restarts",
)

var binPath = flag.String(
	"bin",
	"",
	"directory containing backend-specific scripts (i.e. ./create.sh)",
)

var depotPath = flag.String(
	"depot",
	"",
	"directory in which to store containers",
)

var rootFSPath = flag.String(
	"rootfs",
	"",
	"directory of the rootfs for the containers",
)

var disableQuotas = flag.Bool(
	"disableQuotas",
	false,
	"disable disk quotas",
)

var containerGraceTime = flag.Duration(
	"containerGraceTime",
	0,
	"time after which to destroy idle containers",
)

var debug = flag.Bool(
	"debug",
	false,
	"show low-level command output",
)

func main() {
	flag.Parse()

	maxProcs := runtime.NumCPU()
	prevMaxProcs := runtime.GOMAXPROCS(maxProcs)

	log.Println("set GOMAXPROCS to", maxProcs, "was", prevMaxProcs)

	if *binPath == "" {
		log.Fatalln("must specify -bin with linux backend")
	}

	if *depotPath == "" {
		log.Fatalln("must specify -depot with linux backend")
	}

	if *rootFSPath == "" {
		log.Fatalln("must specify -rootfs with linux backend")
	}

	uidPool := uid_pool.New(10000, 256)

	_, ipNet, err := net.ParseCIDR("10.254.0.0/22")
	if err != nil {
		log.Fatalln("error parsing CIDR:", err)
	}

	networkPool := network_pool.New(ipNet)

	// TODO: base on ephemeral port range
	portPool := port_pool.New(61000, 6501)

	runner := linux_command_runner.New(*debug)

	quotaManager, err := quota_manager.New(*depotPath, *binPath, runner)
	if err != nil {
		log.Fatalln("error creating quota manager:", err)
	}

	if *disableQuotas {
		quotaManager.Disable()
	}

	pool := container_pool.New(
		*binPath,
		*depotPath,
		*rootFSPath,
		uidPool,
		networkPool,
		portPool,
		runner,
		quotaManager,
	)

	backend := linux_backend.New(pool, *snapshotsPath)

	log.Println("setting up backend")

	err = backend.Setup()
	if err != nil {
		log.Fatalln("failed to set up backend:", err)
	}

	log.Println("starting server; listening with", *listenNetwork, "on", *listenAddr)

	graceTime := *containerGraceTime

	wardenServer := server.New(*listenNetwork, *listenAddr, graceTime, backend)

	err = wardenServer.Start()
	if err != nil {
		log.Fatalln("failed to start:", err)
	}

	signals := make(chan os.Signal, 1)

	go func() {
		<-signals
		log.Println("stopping...")
		wardenServer.Stop()
		os.Exit(0)
	}()

	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	select {}
}
