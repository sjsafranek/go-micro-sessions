package api

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/sjsafranek/go-micro-sessions/lib/database"
)

const (
	VERSION = "0.0.1"
)

type Request struct {
	Id      string         `json:"id,omitempty"`
	Version string         `json:"version,omitempty"`
	Method  string         `json:"method,omitempty"`
	Params  *RequestParams `json:"params,omitempty"`
}

type RequestParams struct {
	Email     string           `json:"email,omitempty"`
	Username  string           `json:"username,omitempty"`
	Password  string           `json:"password,omitempty"`
	Apikey    string           `json:"apikey,omitempty"`
	Name      string           `json:"name,omitempty"`
	Timestamp *time.Time       `json:"timestamp,string,omitempty"`
	Filter    *database.Filter `json:"filter,omitempty"`
}

func (self *Request) Unmarshal(data string) error {
	return json.Unmarshal([]byte(data), self)
}

func (self *Request) Marshal() (string, error) {
	b, err := json.Marshal(self)
	if nil != err {
		return "", err
	}
	return string(b), err
}

type Response struct {
	Id      string       `json:"id,omitempty"`
	Version string       `json:"version"`
	Status  string       `json:"status"`
	Message string       `json:"message,omitempty"`
	Error   string       `json:"error,omitempty"`
	Data    ResponseData `json:"data,omitempty"`
}

type ResponseData struct {
	Users []*database.User `json:"users,omitempty"`
	User  *database.User   `json:"user,omitempty"`
}

func (self *Response) Unmarshal(data string) error {
	return json.Unmarshal([]byte(data), self)
}

func (self *Response) Marshal() (string, error) {
	b, err := json.Marshal(self)
	if nil != err {
		return "", err
	}
	return string(b), err
}

func (self *Response) SetError(err error) {
	self.Status = "error"
	self.Error = err.Error()
}

func (self *Response) Write(w io.Writer) error {
	payload, err := self.Marshal()
	if nil != err {
		return err
	}
	_, err = fmt.Fprintln(w, payload)
	return err
}
