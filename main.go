package main

import (
	"crypto/ecdsa"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"wpkg.dev/wpkgup/config"
	"wpkg.dev/wpkgup/crypto"
	"wpkg.dev/wpkgup/server"
	"wpkg.dev/wpkgup/utils"
)

var serverFlag, genFlag, importKeysFlag *flag.FlagSet

func help() {
	fmt.Fprintln(os.Stderr, "\nWPKG toolchain manager")
	fmt.Fprintln(os.Stderr, "\nSubcommands:")
	fmt.Fprintln(os.Stderr, "\ngen-keys - generating keys for client")
	fmt.Fprintln(os.Stderr, "\nimport-keys - import keys for client")
	fmt.Fprintln(os.Stderr, "\nserver - starting server")

	fmt.Fprintln(os.Stderr, "\nServer flags:")
	serverFlag.PrintDefaults()

	fmt.Fprintln(os.Stderr, "\nGen keys flags:")
	genFlag.PrintDefaults()
}

func importKeys(privateKey *ecdsa.PrivateKey, keyringDir string) {
	publicKey := crypto.GeneratePublicFromPrivate(privateKey)

	err := crypto.SavePrivateKeyToFile(privateKey, filepath.Join(keyringDir, "private.pem"))
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	err = crypto.SavePublicKeyToFile(publicKey, filepath.Join(keyringDir, "public.pem"))
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	fmt.Println("Key imported successfully!")
}

func main() {
	gin.SetMode(gin.ReleaseMode)

	var serverIp, workDir string
	var serverPort int

	serverFlag = flag.NewFlagSet("server", flag.ExitOnError)
	serverFlag.StringVar(&serverIp, "i", "0.0.0.0", "Server IP")
	serverFlag.IntVar(&serverPort, "p", 8080, "Server port")
	serverFlag.StringVar(&workDir, "w", config.FindAppDataFolder("wpkgup2"), "Server workdir")

	genFlag = flag.NewFlagSet("gen-keys", flag.ExitOnError)
	genFlag.StringVar(&workDir, "w", config.FindAppDataFolder("wpkgup2"), "Server workdir")

	var keyFile, keyString string

	importKeysFlag = flag.NewFlagSet("import-keys", flag.ExitOnError)
	importKeysFlag.StringVar(&workDir, "w", config.FindAppDataFolder("wpkgup2"), "Server workdir")
	importKeysFlag.StringVar(&keyString, "k", "", "Private key to import")
	importKeysFlag.StringVar(&keyFile, "kf", "", "Private key to import from file")

	println("WpkgUp2", config.Version)

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Expected subcommand")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "server":
		serverFlag.Parse(os.Args[2:])
		config.InitDirs(workDir)

		var conf config.Config
		configFilePath := filepath.Join(workDir, config.ConfigFile)
		if !utils.FileExists(configFilePath) {
			fmt.Println("WpkgUp config not exists... Running configuration...")
			fmt.Print("Set password: ")
			password := utils.ScanDefault("", true)

			conf = config.Config{
				Password: password,
			}
			config.Save(conf, configFilePath)
		}

		server.StartServer(serverIp, serverPort)
	case "gen-keys":
		genFlag.Parse(os.Args[2:])
		config.InitDirs(workDir)

		keyringDir := filepath.Join(config.WorkDir, config.KeyringDir)
		crypto.GenKeys(filepath.Join(keyringDir, "private.pem"), filepath.Join(keyringDir, "public.pem"))
		fmt.Println("Keys generated succesfully")
	case "import-keys":
		importKeysFlag.Parse(os.Args[2:])
		config.InitDirs(workDir)

		if keyFile != "" {
			keyringDir := filepath.Join(config.WorkDir, config.KeyringDir)
			privateKey, err := crypto.ParsePrivateKeyFromFile(keyFile)
			if err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}
			importKeys(privateKey, keyringDir)
		} else if keyString != "" {
			keyringDir := filepath.Join(config.WorkDir, config.KeyringDir)
			privateKey, err := crypto.ParsePrivateKeyFromString(keyString)
			if err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}
			importKeys(privateKey, keyringDir)
		} else {
			fmt.Println("No key specified (-kf or -k option are required)")
			os.Exit(1)
		}

	case "--help":
		help()
	}
}
