package api

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/karlseguin/ccache"
	"github.com/sjsafranek/go-micro-sessions/lib/config"
	"github.com/sjsafranek/go-micro-sessions/lib/database"
	"github.com/sjsafranek/logger"
)

func New(conf *config.Config) *Api {
	return &Api{
		config: conf,
		db:     database.New(conf.Database.GetDatabaseConnection()),
		cache:  ccache.Layered(ccache.Configure()),
	}
}

type Api struct {
	config         *config.Config
	db             *database.Database
	cache          *ccache.LayeredCache
	eventListeners map[string][]func(string, string, float64)
}

func (self *Api) IsPublicMethod(method string) bool {
	return self.config.Api.IsPublicMethod(method)
}

func (self *Api) GetDatabase() *database.Database {
	return self.db
}

func (self *Api) RegisterEventListener(username string, clbk func(string, string, float64)) {
	if nil == self.eventListeners {
		self.eventListeners = make(map[string][]func(string, string, float64))
	}
	self.eventListeners[username] = append(self.eventListeners[username], clbk)
}

func (self *Api) fireEvent(username, device_id, location_id string, probability float64) {
	logger.Info(username, device_id, location_id, probability)
	if clbks, ok := self.eventListeners[username]; ok {
		for i := 0; i < len(clbks); i++ {
			clbks[i](device_id, location_id, probability)
		}
	}
}

func (self *Api) fetchUser(request *Request, clbk func(*database.User) error) error {
	var user *database.User
	var err error
	if "" != request.Params.Apikey {
		user, err = self.getUserByApikey(request.Params.Apikey)
	} else if "" != request.Params.Username {
		user, err = self.getUserByUsername(request.Params.Username)
	} else {
		err = errors.New("Missing parameters")
	}
	if nil != err {
		return err
	}
	return clbk(user)
}

// CreateUser
func (self *Api) createUser(email, username, password string) (*database.User, error) {
	user, err := self.db.CreateUser(email, username)
	if nil == err {
		// cache apikey user pair
		err = user.SetPassword(password)
		if nil == err {
			self.cache.Set("user", user.Apikey, user, 5*time.Minute)
		}
	}
	return user, err
}

// GetUserByUserName
func (self *Api) getUserByUsername(username string) (*database.User, error) {
	return self.db.GetUserByUsername(username)
}

// GetUserByApikey fetches user via apikey. This method uses an inmemory LRU cache to
// decrease the number of database transactions.
func (self *Api) getUserByApikey(apikey string) (*database.User, error) {
	// check cache for apikey user pair
	item := self.cache.Get("user", apikey)
	if nil != item {
		return item.Value().(*database.User), nil
	}

	user, err := self.db.GetUserByApikey(apikey)
	if nil == err {
		// cache apikey user pair
		self.cache.Set("user", apikey, user, 5*time.Minute)
	}
	return user, err
}

func (self *Api) DoJSON(jdata string) (*Response, error) {
	var request Request
	err := json.Unmarshal([]byte(jdata), &request)
	if nil != err {
		response := &Response{Status: "err"}
		response.SetError(err)
		return response, err
	}
	return self.Do(&request)
}

func (self *Api) Do(request *Request) (*Response, error) {
	var response Response

	payload, _ := request.Marshal()
	logger.Debug(payload)

	// TODO HANDLE API VERSIONS
	response.Version = request.Version
	if "" == request.Version {
		response.Version = VERSION
		request.Version = VERSION
	}

	response.Status = "ok"
	response.Id = request.Id

	err := func() error {
		switch request.Method {

		case "get_database_version":
			// {"method":"get_database_version"}
			version, err := self.db.GetVersion()
			if nil != err {
				return err
			}
			response.Message = version
			return nil

		case "ping":
			// {"method":"ping"}
			response.Message = "pong"
			return nil

		case "create_user":
			// {"method":"create_user","params":{"username":"admin_user","email":"admin@email.com","password":"1234"}}
			logger.Info(request)
			if "" == request.Params.Username {
				return errors.New("missing parameters")
			}

			user, err := self.createUser(request.Params.Email, request.Params.Username, request.Params.Password)
			if nil != err {
				return err
			}

			response.Data.User = user
			return nil

		case "get_users":
			// {"method":"get_users"}
			users, err := self.db.GetUsers()
			if nil != err {
				return err
			}
			response.Data.Users = users
			return nil

		case "get_user":
			// {"method":"get_user","params":{"username":"admin_user"}}
			// {"method":"get_user","params":{"apikey":"<apikey>"}}
			if "" == request.Params.Username && "" == request.Params.Apikey {
				return errors.New("missing parameters")
			}
			return self.fetchUser(request, func(user *database.User) error {
				response.Data.User = user
				return nil
			})

		case "delete_user":
			// {"method":"delete_user","username":"admin_user"}
			// {"method":"delete_user","apikey":"<apikey>"}
			return self.fetchUser(request, func(user *database.User) error {
				self.cache.Delete("user", user.Apikey)
				return user.Delete()
			})

		case "activate_user":
			// {"method":"activate_user","username":"admin_user"}
			// {"method":"activate_user","apikey":"<apikey>"}
			return self.fetchUser(request, func(user *database.User) error {
				self.cache.Delete("user", user.Apikey)
				return user.Activate()
			})

		case "deactivate_user":
			// {"method":"deactivate_user","username":"admin_user"}
			// {"method":"deactivate_user","apikey":"<apikey>"}
			return self.fetchUser(request, func(user *database.User) error {
				self.cache.Delete("user", user.Apikey)
				return user.Deactivate()
			})

		case "set_password":
			// {"method":"set_password","username":"admin_user","password":"1234"}
			// {"method":"set_password","apikey":"<apikey>","password":"1234"}
			return self.fetchUser(request, func(user *database.User) error {
				return user.SetPassword(request.Params.Password)
			})

		default:
			return errors.New("method not found")

		}
	}()

	if nil != err {
		response.SetError(err)
	}

	return &response, err
}
