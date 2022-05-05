package route

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
	"net/http"
	"tfacoinlist/response"
)

type manualRegistrationData struct {
	Secret      string `json:"secret"`
	AccountName string `json:"accountName"`
}

func ManualRegistration(db *leveldb.DB) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		var (
			data manualRegistrationData
			key  string
			err  error
		)

		rm := response.NewManager(w, r)

		var b []byte
		if b, err = rm.ReadBody(); err != nil {
			rm.JsonError(http.StatusInternalServerError, "io_read")
			return
		}

		if err = json.Unmarshal(b, &data); err != nil {
			log.WithFields(log.Fields{
				"route":    "ManualRegistration",
				"locality": "json.Unmarshal",
				"body":     string(b),
			}).Error(err)
			rm.JsonError(http.StatusBadRequest, "json_unmarshal")
			return
		}

		if data.Secret == "" || data.AccountName == "" {
			rm.JsonError(http.StatusBadRequest, "bad_data", "Bad data.")
			return
		}

		if key, err = saveNewAccount(db, data.AccountName, data.Secret); err != nil {
			log.WithFields(log.Fields{
				"route":    "ManualRegistration",
				"locality": "saveNewAccount",
			}).Error(err)
			rm.JsonError(http.StatusBadRequest, "save_db")
			return
		}

		sendAccountKey(rm, key)
	}
}

func saveNewAccount(db *leveldb.DB, accountName, secret string) (key string, err error) {
	h := sha1.New()
	h.Write([]byte(accountName + secret))
	key = fmt.Sprintf("%x", h.Sum(nil))

	err = db.Put([]byte(accountName), []byte(secret+":"+key), nil)
	return
}

func sendAccountKey(rm *response.Manager, key string) {
	resData := struct {
		Key string `json:"key"`
	}{
		Key: key,
	}

	rm.Json(resData, http.StatusOK, nil)
}
