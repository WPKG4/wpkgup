package main

import (
	"crypto/ecdsa"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"wpkg.dev/wpkgup/client"
	"wpkg.dev/wpkgup/config"
	"wpkg.dev/wpkgup/crypto"
	"wpkg.dev/wpkgup/keystore"
	"wpkg.dev/wpkgup/server"
	"wpkg.dev/wpkgup/utils"
)

var initFlag, serverFlag, genFlag, importKeysFlag, uploadKeysFlag, signBinaryFlag, uploadBinaryFlag *flag.FlagSet

func help(argv0 string) {
	fmt.Fprintln(os.Stderr, "\nWPKG update manager")
	fmt.Fprintln(os.Stderr, "\nUsage: "+argv0+" <command> [command options]")
	fmt.Fprintln(os.Stderr, "\nCommands:")
	fmt.Fprintln(os.Stderr, "\nserver - starting server")
	serverFlag.PrintDefaults()
	fmt.Fprintln(os.Stderr, "\ngen-keys - generating keys for client")
	genFlag.PrintDefaults()
	fmt.Fprintln(os.Stderr, "\nimport-keys - import keys for client")
	importKeysFlag.PrintDefaults()
	fmt.Fprintln(os.Stderr, "\nupload-keys - upload keys to server")
	uploadKeysFlag.PrintDefaults()
	fmt.Fprintln(os.Stderr, "\nsign-binary <binary to sign> <sign file output> - Sign binary")
	fmt.Fprintln(os.Stderr, "\nupload-binary <component> <channel> <os> <arch> <version> <filename> - Upload binary to server binary")
	uploadBinaryFlag.PrintDefaults()
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

	initFlag = flag.NewFlagSet("init", flag.ExitOnError)
	initFlag.StringVar(&workDir, "w", config.FindAppDataFolder("wpkgup2"), "Server workdir")

	var keyFile, keyString string

	importKeysFlag = flag.NewFlagSet("import-keys", flag.ExitOnError)
	importKeysFlag.StringVar(&workDir, "w", config.FindAppDataFolder("wpkgup2"), "Server workdir")
	importKeysFlag.StringVar(&keyString, "k", "", "Private key to import")
	importKeysFlag.StringVar(&keyFile, "kf", "", "Private key to import from file")

	var address, password string

	uploadKeysFlag = flag.NewFlagSet("upload-keys", flag.ExitOnError)
	uploadKeysFlag.StringVar(&address, "i", "0.0.0.0:8080", "Server Address")
	uploadKeysFlag.StringVar(&password, "p", "", "Server Password")
	uploadKeysFlag.StringVar(&workDir, "w", config.FindAppDataFolder("wpkgup2"), "Server workdir")

	signBinaryFlag = flag.NewFlagSet("sign-binary", flag.ExitOnError)
	signBinaryFlag.StringVar(&workDir, "w", config.FindAppDataFolder("wpkgup2"), "Server workdir")

	uploadBinaryFlag = flag.NewFlagSet("upload-binary", flag.ExitOnError)
	uploadBinaryFlag.StringVar(&address, "i", "0.0.0.0:8080", "Server Address")
	uploadBinaryFlag.StringVar(&workDir, "w", config.FindAppDataFolder("wpkgup2"), "Server workdir")
	uploadBinaryFlag.StringVar(&keyString, "k", "", "Private key to import")

	println("WpkgUp2", config.Version)

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Expected command, use --help argument to print help")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "init":
		initFlag.Parse(os.Args[2:])
		config.InitDirs(workDir)

		var conf config.Config
		configFilePath := filepath.Join(workDir, config.ConfigFile)
		if !utils.FileExists(configFilePath) {
			fmt.Println("Running configuration...")
			fmt.Print("Set password: ")
			password := utils.ScanRequired()

			conf = config.Config{
				Password: password,
			}
			config.Save(conf, configFilePath)

			fmt.Println("Config created!")
		} else {
			fmt.Println("Config already exists in this directory")
			os.Exit(1)
		}
	case "server":
		serverFlag.Parse(os.Args[2:])
		config.InitDirs(workDir)
		err := keystore.Init()
		if err != nil {
			fmt.Println("Failed to init keystore:", err)
			os.Exit(1)
		}

		var conf config.Config
		configFilePath := filepath.Join(workDir, config.ConfigFile)

		if !utils.FileExists(configFilePath) {
			defaultPassword := "1@Qwerty"

			fmt.Println("Config not detected, creating default config...")
			fmt.Println("Default password is \"" + defaultPassword + "\", remember to change it later.")
			conf = config.Config{
				Password: defaultPassword,
			}
			err := config.Save(conf, configFilePath)
			if err != nil {
				fmt.Println("Failed to save config")
				os.Exit(1)
			}
		}

		err = config.Init()
		if err != nil {
			fmt.Println("Failed to load config")
			os.Exit(1)
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
	case "upload-keys":
		uploadKeysFlag.Parse(os.Args[2:])
		config.InitDirs(workDir)

		if password == "" {
			fmt.Print("Enter server password: ")
			password = utils.ScanRequired()
		}

		err := client.UploadKey(address, password)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println("Keys uploaded successfully!")
	case "sign-binary":
		signBinaryFlag.Parse(os.Args[2:])
		config.InitDirs(workDir)

		binaryFilePath := os.Args[2]
		signFilePath := os.Args[3]

		privateKey, err := crypto.ParsePrivateKeyFromFile(filepath.Join(config.WorkDir, config.KeyringDir, "private.pem"))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		signBuffer, err := crypto.Sign(privateKey, binaryFilePath)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		err = os.WriteFile(signFilePath, signBuffer, 0664)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	case "upload-binary":
		uploadBinaryFlag.Parse(os.Args[7:])
		config.InitDirs(workDir)

		component := os.Args[2]
		channel := os.Args[3]
		Os := os.Args[4]
		arch := os.Args[5]
		version := os.Args[6]
		filename := os.Args[7]

		var privateKey *ecdsa.PrivateKey
		var err error

		fmt.Println(keyString)

		if keyString != "" {
			privateKey, err = crypto.ParsePrivateKeyFromString(keyString)
		} else {
			privateKey, err = crypto.ParsePrivateKeyFromFile(filepath.Join(config.WorkDir, config.KeyringDir, "private.pem"))
		}
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		err = client.UploadBinary(component, channel, Os, arch, version, address, filename, privateKey)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println("Binary uploaded successfully!")
	case "--help":
		help(os.Args[0])
	}
}
