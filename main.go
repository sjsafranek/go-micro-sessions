package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/sjsafranek/go-micro-sessions/lib/api"
	"github.com/sjsafranek/go-micro-sessions/lib/clients/repl"
	"github.com/sjsafranek/go-micro-sessions/lib/clients/web"
	"github.com/sjsafranek/go-micro-sessions/lib/config"
	"github.com/sjsafranek/logger"
)

const (
	PROJECT                   string = "Go Micro Sessions"
	VERSION                   string = "0.0.1"
	DEFAULT_HTTP_PORT         int    = 8080
	DEFAULT_DATABASE_ENGINE   string = "postgres"
	DEFAULT_DATABASE_DATABASE string = "sessionsdb"
	DEFAULT_DATABASE_PASSWORD string = "dev"
	DEFAULT_DATABASE_USERNAME string = "sessionsuser"
	DEFAULT_DATABASE_HOST     string = "localhost"
	DEFAULT_DATABASE_PORT     int64  = 5432
	DEFAULT_REDIS_PORT        int64  = 6379
	DEFAULT_REDIS_HOST        string = ""
	DEFAULT_CONFIG_FILE       string = "config.json"
)

var (
	HTTP_PORT              int    = DEFAULT_HTTP_PORT
	FACEBOOK_CLIENT_ID     string = os.Getenv("FACEBOOK_CLIENT_ID")
	FACEBOOK_CLIENT_SECRET string = os.Getenv("FACEBOOK_CLIENT_SECRET")
	GOOGLE_CLIENT_ID       string = os.Getenv("GOOGLE_CLIENT_ID")
	GOOGLE_CLIENT_SECRET   string = os.Getenv("GOOGLE_CLIENT_SECRET")
	DATABASE_ENGINE        string = DEFAULT_DATABASE_ENGINE
	DATABASE_DATABASE      string = DEFAULT_DATABASE_DATABASE
	DATABASE_PASSWORD      string = DEFAULT_DATABASE_PASSWORD
	DATABASE_USERNAME      string = DEFAULT_DATABASE_USERNAME
	DATABASE_HOST          string = DEFAULT_DATABASE_HOST
	DATABASE_PORT          int64  = DEFAULT_DATABASE_PORT
	REDIS_PORT             int64  = DEFAULT_REDIS_PORT
	REDIS_HOST             string = DEFAULT_REDIS_HOST
	CONFIG_FILE            string = DEFAULT_CONFIG_FILE
	API_REQUEST            string = ""
	MODE                   string = "web"
	API               	   *api.Api
	conf                   *config.Config
)

func init() {
	var printVersion bool

	// read credentials from environment variables if available
	conf = &config.Config{
		Api: config.Api{
			PublicMethods: []string{
				"set_password",
			},
		},
		Server: config.Server{
			HttpPort: DEFAULT_HTTP_PORT,
		},
		OAuth2: config.OAuth2{
			Facebook: config.SocialOAuth2{
				ClientID:     FACEBOOK_CLIENT_ID,
				ClientSecret: FACEBOOK_CLIENT_SECRET,
			},
			Google: config.SocialOAuth2{
				ClientID:     GOOGLE_CLIENT_ID,
				ClientSecret: GOOGLE_CLIENT_SECRET,
			},
		},
		Database: config.Database{
			DatabaseEngine: DATABASE_ENGINE,
			DatabaseHost:   DEFAULT_DATABASE_HOST,
			DatabaseName:   DEFAULT_DATABASE_DATABASE,
			DatabasePass:   DEFAULT_DATABASE_PASSWORD,
			DatabaseUser:   DEFAULT_DATABASE_USERNAME,
			DatabasePort:   DEFAULT_DATABASE_PORT,
		},
		Redis: config.Redis{
			Host: DEFAULT_REDIS_HOST,
			Port: DEFAULT_REDIS_PORT,
		},
	}

	flag.IntVar(&conf.Server.HttpPort, "httpport", DEFAULT_HTTP_PORT, "Server port")
	flag.StringVar(&conf.OAuth2.Facebook.ClientID, "facebook-client-id", FACEBOOK_CLIENT_ID, "Facebook Client ID")
	flag.StringVar(&conf.OAuth2.Facebook.ClientSecret, "facebook-client-secret", FACEBOOK_CLIENT_SECRET, "Facebook Client Secret")
	flag.StringVar(&conf.OAuth2.Google.ClientID, "gmail-client-id", GOOGLE_CLIENT_ID, "Google Client ID")
	flag.StringVar(&conf.OAuth2.Google.ClientSecret, "gmail-client-secret", GOOGLE_CLIENT_SECRET, "Google Client Secret")
	flag.StringVar(&conf.Database.DatabaseHost, "dbhost", DEFAULT_DATABASE_HOST, "database host")
	flag.StringVar(&conf.Database.DatabaseName, "dbname", DEFAULT_DATABASE_DATABASE, "database name")
	flag.StringVar(&conf.Database.DatabasePass, "dbpass", DEFAULT_DATABASE_PASSWORD, "database password")
	flag.StringVar(&conf.Database.DatabaseUser, "dbuser", DEFAULT_DATABASE_USERNAME, "database username")
	flag.Int64Var(&conf.Database.DatabasePort, "dbport", DEFAULT_DATABASE_PORT, "Database port")

	flag.StringVar(&conf.Redis.Host, "redishost", DEFAULT_REDIS_HOST, "Redis host")
	flag.Int64Var(&conf.Redis.Port, "redisport", DEFAULT_REDIS_PORT, "Redis port")

	flag.StringVar(&CONFIG_FILE, "c", DEFAULT_CONFIG_FILE, "config file")

	flag.StringVar(&API_REQUEST, "query", "", "Api query to execute")
	flag.BoolVar(&printVersion, "V", false, "Print version and exit")
	flag.Parse()

	if printVersion {
		fmt.Println(PROJECT, VERSION)
		os.Exit(0)
	}

	args := flag.Args()
	if 1 == len(args) {
		MODE = args[0]
	}

}

func main() {

	API = api.New(conf)

	if "" != API_REQUEST {
		request := api.Request{}
		request.Unmarshal(API_REQUEST)
		response, err := API.Do(&request)
		if nil != err {
			panic(err)
		}

		results, err := response.Marshal()
		if nil != err {
			panic(err)
		}
		fmt.Println(results)
		return
	}

	logger.Debug("GOOS: ", runtime.GOOS)
	logger.Debug("CPUS: ", runtime.NumCPU())
	logger.Debug("PID: ", os.Getpid())
	logger.Debug("Go Version: ", runtime.Version())
	logger.Debug("Go Arch: ", runtime.GOARCH)
	logger.Debug("Go Compiler: ", runtime.Compiler)
	logger.Debug("NumGoroutine: ", runtime.NumGoroutine())

	resp, err := API.DoJSON(`{"method":"get_database_version"}`)
	if nil != err {
		panic(err)
	}
	logger.Debugf("Database version: %v", resp.Message)

	switch MODE {

	case "repl":
		repl.New(API).Run()
		break

	case "web":
		app := web.New(API, conf)
		err = app.ListenAndServe(fmt.Sprintf(":%v", conf.Server.HttpPort))
		if err != nil {
			panic(err)
		}
		break

	default:
		panic(errors.New("api client not found"))
	}

}
