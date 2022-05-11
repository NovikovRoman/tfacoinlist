package route

import (
	"crypto/sha1"
	"fmt"
	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"tfacoinlist/response"
	"time"
)

func ManualRegistrationGET() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		outPageManualRegistration(response.NewManager(w, r), "")
	}
}

func ManualRegistration(db *leveldb.DB) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var (
			key    string
			secret string
			values url.Values
			qr     qrCode
			err    error
		)

		rm := response.NewManager(w, r)

		var b []byte
		if b, err = rm.ReadBody(); err != nil || len(b) == 0 {
			rm.JsonError(http.StatusInternalServerError, "io_read")
			return
		}

		if values, err = url.ParseQuery(string(b)); err != nil {
			log.WithFields(log.Fields{
				"route":    "ManualRegistration",
				"locality": "ParseQuery",
				"query":    string(b),
			}).Error(err)
			rm.JsonError(http.StatusInternalServerError, "parse_url")
			return
		}

		accountName := values.Get("accountName")
		qrcodeUrl := values.Get("qrcodeUrl")

		// Чистый секретный ключ
		if regexp.MustCompile(`(?si)^[0-9a-z]+$`).MatchString(strings.ToLower(qrcodeUrl)) {
			secret = qrcodeUrl

			// Ссылка otpauth
		} else if strings.Contains(qrcodeUrl, "otpauth://") {
			var u *url.URL

			if u, err = url.Parse(qrcodeUrl); err != nil {
				log.WithFields(log.Fields{
					"route":    "ManualRegistration",
					"locality": "url.Parse",
					"url":      qrcodeUrl,
				}).Error(err)
				rm.JsonError(http.StatusBadRequest, "otpauth_parse")
				return
			}

			/*ar := strings.Split(strings.TrimLeft(u.Path, "/"), ":")
			accountName = ar[0]
			if len(ar) == 2 {
				accountName = ar[1]
			}*/
			secret = u.Query().Get("secret")

			// Ссылка на qrcode
		} else {
			if qr, err = getQrCodeByUrl(qrcodeUrl); err != nil {
				log.WithFields(log.Fields{
					"route":    "ManualRegistration",
					"locality": "getQrCodeByUrl",
					"qrCode":   qrcodeUrl,
				}).Error(err)
				rm.JsonError(http.StatusBadRequest, "download_qrcode")
				return
			}

			accountName = qr.accountName
			secret = qr.secret
		}

		if key, err = saveNewAccount(db, accountName, secret); err != nil {
			log.WithFields(log.Fields{
				"route":    "ManualRegistration",
				"locality": "saveNewAccount",
			}).Error(err)
			rm.JsonError(http.StatusBadRequest, "save_db")
			return
		}

		header := fmt.Sprintf(
			`<h4>Зарегистрирован.</h4><p><a target="_blank" href="/auth/totp/%s/%s/">сгенерированный код</a> обновляется каждые 30 секунд</p><p>key: %s</p><hr>`,
			accountName, key, key,
		)
		outPageManualRegistration(rm, header)
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

func getQrCodeByUrl(qrcodeUrl string) (qr qrCode, err error) {
	var b []byte
	if b, err = downloadFile(qrcodeUrl, time.Second*30); err != nil {
		return
	}

	qr, err = qrCodeFromByte(b)
	return
}

func outPageManualRegistration(rm *response.Manager, header string) {
	var err error

	page := `<html><head></head><body>` + header + `<h4>Регистрация</h4><form method="post" style="width:300px;">
<p><label for="inputAccount">Аккаунт (email)<br>
<input id="inputAccount" type="text" name="accountName" style="width:100%;" /></p>
<p><label for="inputQrcodeUrl">Ссылка на картинку qrcode, секретный код или ссылку otpauth://…<br>
<input id="inputQrcodeUrl" type="text" name="qrcodeUrl" style="width:100%;" /></label></p>
<p><button type="submit">Отправить</button></p>
</form></body></html>`

	if _, err = rm.ResponseWriter.Write([]byte(page)); err != nil {
		rm.ResponseWriter.WriteHeader(http.StatusInternalServerError)
		log.WithFields(log.Fields{
			"locality": "ManualRegistrationGET",
		}).Error(err)
	}
}
