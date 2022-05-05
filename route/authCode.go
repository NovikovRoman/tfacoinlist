package route

import (
	"bytes"
	"github.com/julienschmidt/httprouter"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	log "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
	"net/http"
	"strings"
	"tfacoinlist/response"
	"time"
)

const (
	period = 30
)

func AuthCode(db *leveldb.DB) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var (
			value       []byte
			ar          [][]byte
			key         string
			accountName string
			code        string
			err         error
		)

		rm := response.NewManager(w, r)

		key = strings.ToLower(p.ByName("key"))
		accountName = strings.ToLower(p.ByName("email"))

		if value, err = db.Get([]byte(accountName), nil); err != nil && err != leveldb.ErrNotFound {
			log.WithFields(log.Fields{
				"route":       "AuthCode",
				"locality":    "db.Get",
				"accountName": accountName,
			}).Error(err)
			rm.JsonError(http.StatusBadRequest, "db_get")
			return
		}

		if err == leveldb.ErrNotFound {
			rm.JsonError(http.StatusBadRequest, "user_not_found")
		}

		if ar = bytes.Split(value, []byte(":")); len(ar) != 2 {
			log.WithFields(log.Fields{
				"route":    "AuthCode",
				"locality": "bytes.Split",
				"value":    string(value),
			}).Error(err)
			rm.JsonError(http.StatusBadRequest, "bad_db_data")
			return
		}

		if string(ar[1]) != key {
			rm.Json("", http.StatusUnauthorized, nil)
			return
		}

		code, err = totp.GenerateCodeCustom(
			string(ar[0]),
			time.Now().UTC(),
			totp.ValidateOpts{
				Period: uint(period),
				Digits: otp.DigitsSix,
			},
		)
		if err != nil {
			log.WithFields(log.Fields{
				"route":    "AuthCode",
				"locality": "totp.GenerateCodeCustom",
			}).Error(err)
			rm.JsonError(http.StatusBadRequest, "generate_code")
			return
		}

		resData := struct {
			Code string `json:"code"`
		}{
			Code: code,
		}

		rm.Json(resData, http.StatusOK, nil)
	}
}
