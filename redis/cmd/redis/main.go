package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"redis/iriscore/config"
	"redis/iriscore/iocgo"
	"redis/iriscore/router"
	"redis/iriscore/version"

	"gitlab.10101111.com/oped/DBMS_LIBS/debug"
	"gitlab.10101111.com/oped/DBMS_LIBS/perf"

	log "gitlab.10101111.com/oped/DBMS_LIBS/logrus"

	// register ioc
	_ "redis/iriscore/middleware/tracinglog"
	_ "redis/iriscore/resource"
	_ "redis/iriscore/service"
	_ "redis/thirdurl"
)

const (
	devops      = `redis`
	HttpTimeout = 30
)

var cfgfile = flag.String("c", "./data/conf/config.toml", "configuration file, default to config.toml")
var ver = flag.Bool("version", false, "Output version and exit")

//var Service string //= "devops"
//var Version string //= "no_version" /* Version for dbms main.version*/

func main() {
	// args parse
	flag.Parse()

	//version.Service = Service
	//version.Version = Version
	version.Service = devops

	if *ver {
		fmt.Println(version.Service, ": ", version.Version)
		return
	}

	// config file parse
	cfg, err := config.NewConfig(*cfgfile)
	if err != nil {
		fmt.Printf(" load conf %s err:%v", *cfgfile, err)
		return
	}

	// logger
	rf := log.NewRotateFile(cfg.Log.Logpath, 100*log.MiB)
	defer rf.Close()
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(rf)

	//  log level
	log.SetLevel(log.DebugLevel)
	switch cfg.Log.Loglevel {
	case "ERROR":
		fallthrough
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "warn":
		fallthrough
	case "WARN":
		log.SetLevel(log.WarnLevel)
	case "INFO":
		fallthrough
	case "info":
		log.SetLevel(log.InfoLevel)
	case "DEBUG":
		fallthrough
	case "debug":
		log.SetLevel(log.DebugLevel)
	}

	//pid file
	err = InitPidfile(cfg)
	if err != nil {
		fmt.Printf("initpidfile err:%v", err)
		return

	}
	defer QuitPidFile(cfg)

	//////////////////////////////////////
	//     service init
	//  init the dependency service
	err = InitDependencyService(cfg)
	if err != nil {
		log.Errorf("init dependency service err:%v", err)
		return
	}
	log.Infof("init dependency service ok")

	//  when service stopping, close the dependency service
	defer CloseDependencyService()

	///////////////////////////////////////
	////  http service
	// perf service
	perf.Init(cfg.Http.PprofAddr)
	log.Infof("http pprof service init ok")

	if cfg.Http.Timeout == 0 {
		cfg.Http.Timeout = HttpTimeout
	}

	// http service init
	router.Api(). // singleTon api
			ConfigDefault().
			SetTimeout(time.Duration(cfg.Http.Timeout) * time.Second).
			SetLog(rf).
			InitRouter().
			Runapi(cfg.Http.HttpAddr)
	log.Infof("http service init ok")

	////////
	log.Infof("          ________                                                     ")
	log.Infof("       __/_/      |______   %s.%s is running                ", version.Service, version.Version)
	log.Infof("      / O O O O O O O O O ...........................................  ")
	log.Infof("                                                                       ")
	log.Infof("      %s", time.Now().String())
	log.Infof("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")

	////////
	// signal
	InitSignal()
}

func InitSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT, syscall.SIGSTOP, syscall.SIGUSR1, syscall.SIGUSR2)
	//log.Infof("ait for signal.......")
	for {
		s := <-c
		log.Infof("service[%s] get a signal %s", version.Version, s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGSTOP, syscall.SIGINT, syscall.SIGHUP:
			GracefulQuit()
			return
		case syscall.SIGUSR2:
			debug.DumpStacks()
		case syscall.SIGUSR1:
			// todo: your operation
			//return
		default:
			//return
		}
	}
}

func GracefulQuit() {
	log.Infof("service make a graceful quit !!!!!!!!!!!!!!")
	router.Api().Shutdown() // close http service
	// close your service here

	time.Sleep(1 * time.Second)
}

func InitDependencyService(cfg *config.ApiConf) error {
	return iocgo.LaunchEngine(cfg)
}

func CloseDependencyService() error {
	return iocgo.StopEngine()
}

func InitPidfile(cfg *config.ApiConf) error {
	//pid file
	if cfg.Log.Pidfile == "" {
		return nil
	}
	contents, err := ioutil.ReadFile(cfg.Log.Pidfile)
	if err == nil {
		pid, err := strconv.Atoi(strings.TrimSpace(string(contents)))
		if err != nil {
			log.Errorf("Error reading proccess id from pidfile '%s': %s",
				cfg.Log.Pidfile, err)
			return err
		}

		process, err := os.FindProcess(pid)

		// on Windows, err != nil if the process cannot be found
		if runtime.GOOS == "windows" {
			if err == nil {
				log.Errorf("Process %d is already running.", pid)
				return fmt.Errorf("already running")
			}
		} else if process != nil {
			// err is always nil on POSIX, so we have to send the process
			// a signal to check whether it exists
			if err = process.Signal(syscall.Signal(0)); err == nil {
				log.Errorf("Process %d is already running.", pid)
				return fmt.Errorf("already running")
			}
		}
	}
	if err = ioutil.WriteFile(cfg.Log.Pidfile, []byte(strconv.Itoa(os.Getpid())),
		0644); err != nil {

		log.Errorf("Unable to write pidfile '%s': %s", cfg.Log.Pidfile, err)
		return err
	}
	log.Infof("Wrote pid to pidfile '%s'", cfg.Log.Pidfile)
	return nil
}

func QuitPidFile(cfg *config.ApiConf) {
	if cfg.Log.Pidfile == "" {
		return
	}
	if err := os.Remove(cfg.Log.Pidfile); err != nil {
		log.Errorf("Unable to remove pidfile '%s': %s", cfg.Log.Pidfile, err)
	}
	return
}
