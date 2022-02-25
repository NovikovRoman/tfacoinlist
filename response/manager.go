package response

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strings"
)

type Manager struct {
	ResponseWriter http.ResponseWriter
	Request        *http.Request
}

func NewManager(w http.ResponseWriter, r *http.Request) *Manager {
	return &Manager{
		ResponseWriter: w,
		Request:        r,
	}
}

func (m Manager) Json(data interface{}, httpStatus int, headers map[string]string) {
	var (
		js  []byte
		err error
	)

	if js, err = json.Marshal(data); err != nil {
		m.ResponseWriter.WriteHeader(http.StatusInternalServerError)
		log.WithFields(log.Fields{
			"locality": "manager.Json json.Marshal",
			"data":     data,
		}).Error(err)
		return
	}

	m.ResponseWriter.Header().Set("Content-Type", "application/json; charset=UTF-8")
	m.ResponseWriter.WriteHeader(httpStatus)
	if headers != nil {
		for key, value := range headers {
			m.ResponseWriter.Header().Set(key, value)
		}
	}

	if _, err = m.ResponseWriter.Write(js); err != nil {
		m.ResponseWriter.WriteHeader(http.StatusInternalServerError)
		log.WithFields(log.Fields{
			"locality": "manager.Json ResponseWriter.Write",
			"js":       js,
		}).Error(err)
		return
	}
}

func (m Manager) JsonError(httpStatus int, code string, messages ...string) {
	if len(messages) == 0 {
		messages = []string{"Ошибка сервера. Обратитесь к разработчику."}
	}

	if code == "" {
		code = "unknown"
	}

	data := struct {
		Message string `json:"msg"`
		Code    string `json:"code"`
	}{
		Message: strings.Join(messages, " "),
		Code:    code,
	}

	m.Json(data, httpStatus, nil)
}

func (m Manager) ReadBody() (b []byte, err error) {
	if b, err = ioutil.ReadAll(m.Request.Body); err != nil {
		log.WithFields(log.Fields{
			"locality": "manager.ReadBody ioutil.ReadAll",
		}).Error(err)
	}
	return
}
