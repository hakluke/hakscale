package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/go-redis/redis"
	"gopkg.in/yaml.v2"
)

//Config struct holds all configuration data that comes from config.yml or environment variables
type Config struct {
	Redis struct {
		Host     string `yaml:"host" envconfig:"REDIS_HOST"`
		Port     string `yaml:"port" envconfig:"REDIS_PORT"`
		Password string `yaml:"password" envconfig:"REDIS_PASSWORD"`
	} `yaml:"redis"`
}

// Global variables
var config *Config
var redisClient *redis.Client

func main() {

	// if less than 2 arguments, error out
	if len(os.Args) < 2 {
		fmt.Println("Error: Subcommand missing or incorrect.\n\nHint: you can push jobs to the queue with:\n\nhakscale push -p \"param1:./file1.txt,param2:./file2.txt\" -c \"nmap -A _param1_ _param2_\" -t 20\n\nOr you can pop them from the queue and execute them with:\n\nhakscale pop -q nmap -t 20\n\nFor full usage instructions, see github.com/hakluke/hakscale")
		return
	}

	// load config file
	f, err := os.Open(os.Getenv("HOME") + "/.config/haktools/hakscale-config.yml")
	if err != nil {
		fmt.Println("Error opening config file:", err)
	}
	defer f.Close()

	// parse the config file
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&config)
	if err != nil {
		fmt.Println("Error decoding config.yml", err)
		return
	}

	// Connect to redis server
	redisClient = redis.NewClient(&redis.Options{
		Addr:      config.Redis.Host + ":" + config.Redis.Port,
		Password:  config.Redis.Password,
		DB:        0,                                     // redis databases are deprecated so we will just use the default
		TLSConfig: &tls.Config{InsecureSkipVerify: true}, // needed for the standard DO redis
	})

	// Check redis server connection
	_, err = redisClient.Ping().Result()
	if err != nil {
		fmt.Println("Unable to connect to specified Redis server:", err)
		os.Exit(1)
	}

	switch os.Args[1] {

	case "push":
		flagSet := flag.NewFlagSet("hakscale push", flag.ExitOnError)
		verbose := flagSet.Bool("v", false, "verbose mode")
		command := flagSet.String("c", "", "the command you wish to scale, including placeholders")
		queue := flagSet.String("q", "cmd", "the name of the queue that you would like to push jobs to")
		parametersString := flagSet.String("p", "", "the placeholders and files being used")
		test := flagSet.Bool("test", false, "print the commands to terminal, don't actually push them to redis")
		timeout := flagSet.Int("t", 0, "timeout for the commands (in seconds)")
		flagSet.Parse(os.Args[2:])
		if *timeout == 0 {
			log.Fatal("You must specify a timeout to avoid leaving your workers endlessly working. Hint: -t <seconds>")
		}
		pushIt(*command, *queue, *parametersString, *test, *timeout, *verbose)
	case "pop":
		flagSet := flag.NewFlagSet("hakscale pop", flag.ExitOnError)
		verbose := flagSet.Bool("v", false, "verbose mode")
		queue := flagSet.String("q", "cmd", "the name of the queue that you would like to pop jobs from")
		threads := flagSet.Int("t", 5, "number of threads")
		flagSet.Parse(os.Args[2:])
		popIt(*threads, *queue, *verbose)

	// no valid subcommand found - default to showing a message and exiting
	default:
		fmt.Println("Error: Subcommand missing or incorrect. Hint, you can push jobs to the queue with:\n\nhakscale push -p \"param1:./file1.txt,param2:./file2.txt\" -c \"nmap -A _param1_ _param2_\" -t 20\n\nOr you can pop them from the queue and execute them with:\n\nhakscale pop -q nmap -t 20")
		os.Exit(1)
	}
}
