package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"wpkg.dev/wpkgup/config"
	"wpkg.dev/wpkgup/server"
)

var serverFlag *flag.FlagSet

func help() {
	fmt.Fprintln(os.Stderr, "\nWPKG toolchain manager")
	fmt.Fprintln(os.Stderr, "\nSubcommands:")
	fmt.Fprintln(os.Stderr, "\nserver - starting server")

	fmt.Fprintln(os.Stderr, "\nServer flags:")
	serverFlag.PrintDefaults()
}

func startServer(ip string, port int) {
	r := gin.Default()
	server.InitControllers(r)
	fmt.Println("Starting HTTP Server at http://" + ip + ":" + strconv.Itoa(port))
	r.Run(ip + ":" + strconv.Itoa(port))
}

func main() {
	gin.SetMode(gin.ReleaseMode)

	var serverIp, workDir string
	var serverPort int

	serverFlag = flag.NewFlagSet("server", flag.ExitOnError)
	serverFlag.StringVar(&serverIp, "i", "0.0.0.0", "Server IP")
	serverFlag.IntVar(&serverPort, "p", 8080, "Server port")
	serverFlag.StringVar(&workDir, "w", config.FindAppDataFolder("wpkgup2"), "Server workdir")

	println("WpkgUp2", config.Version)

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Expected subcommand")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "server":
		serverFlag.Parse(os.Args[2:])
		config.InitDirs(workDir)
		startServer(serverIp, serverPort)
	case "--help":
		help()
		break
	}
}
