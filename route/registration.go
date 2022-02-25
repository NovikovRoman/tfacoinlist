package route

import (
	"bytes"
	"crypto/sha1"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/NovikovRoman/qrcode"
	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"tfacoinlist/response"
	"time"
)

type qrCode struct {
	issuer      string
	accountName string
	secret      string
	period      int
}

type registrationData struct {
	QrCode    string `json:"qrCode"`
	QrCodeUrl string `json:"qrCodeUrl"`
}

func Registration(db *leveldb.DB) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		var (
			b    []byte
			data registrationData
			qr   qrCode
			err  error
		)

		rm := response.NewManager(w, r)

		if data, err = getQrCode(rm); err != nil {
			return
		}

		if data.QrCodeUrl != "" {
			if b, err = downloadFile(data.QrCodeUrl, time.Second*30); err != nil {
				log.WithFields(log.Fields{
					"route":    "Registration",
					"locality": "downloadFile",
					"qrCode":   data.QrCodeUrl,
				}).Error(err)
				rm.JsonError(http.StatusBadRequest, "download_qrcode")
				return
			}

		} else if b, err = base64.StdEncoding.DecodeString(data.QrCode); err != nil {
			log.WithFields(log.Fields{
				"route":    "Registration",
				"locality": "DecodeString",
				"qrCode":   data.QrCode,
			}).Error(err)
			rm.JsonError(http.StatusBadRequest, "decoding_base64_qrcode")
			return
		}

		if qr, err = qrCodeFromByte(b); err != nil {
			log.WithFields(log.Fields{
				"route":      "Registration",
				"locality":   "qrCodeFromByte",
				"qrCodeByte": b,
			}).Error(err)
			rm.JsonError(http.StatusBadRequest, "decoding_qrcode_from_byte")
			return
		}

		h := sha1.New()
		h.Write([]byte(qr.accountName + qr.secret))
		key := fmt.Sprintf("%x", h.Sum(nil))

		if err = db.Put([]byte(qr.accountName), []byte(qr.secret+":"+key), nil); err != nil {
			log.WithFields(log.Fields{
				"route":    "Registration",
				"locality": "db.Put",
			}).Error(err)
			rm.JsonError(http.StatusBadRequest, "save_db")
			return
		}

		resData := struct {
			Key string `json:"key"`
		}{
			Key: key,
		}

		rm.Json(resData, http.StatusOK, nil)
	}
}

// getQrCode получает qrcode для регистрации
func getQrCode(rm *response.Manager) (data registrationData, err error) {
	var b []byte
	if b, err = rm.ReadBody(); err != nil {
		rm.JsonError(http.StatusInternalServerError, "io_read")
		return
	}

	if err = json.Unmarshal(b, &data); err != nil {
		log.WithFields(log.Fields{
			"route":    "Registration",
			"method":   "getQrCode",
			"locality": "json.Unmarshal",
			"body":     string(b),
		}).Error(err)
		rm.JsonError(http.StatusBadRequest, "json_unmarshal")
	}

	return
}

func qrCodeFromByte(b []byte) (qr qrCode, err error) {
	qr.issuer, qr.accountName, qr.secret, qr.period, err = qrCodeDecode(b)
	return
}

func qrCodeDecode(img []byte) (issuer, accountName, secret string, period int, err error) {
	var (
		matrix *qrcode.Matrix
		u      *url.URL
	)

	if matrix, err = qrcode.Decode(bytes.NewReader(img)); err != nil {
		return
	}

	if u, err = url.Parse(matrix.Content); err != nil {
		return
	}

	issuer = u.Query().Get("issuer")
	ar := strings.Split(strings.TrimLeft(u.Path, "/"), ":")
	accountName = ar[0]
	if len(ar) == 2 {
		accountName = ar[1]
	}
	secret = u.Query().Get("secret")

	if u.Query().Get("period") != "" {
		if period, err = strconv.Atoi(u.Query().Get("period")); err != nil {
			return
		}
	}

	if period < 1 {
		period = 30
	}
	return
}

func downloadFile(link string, timeout time.Duration) (b []byte, err error) {
	var (
		resp       *http.Response
		statusCode int
	)

	transport := &http.Transport{
		TLSHandshakeTimeout: timeout,
		IdleConnTimeout:     timeout,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
	}

	client := http.Client{
		Transport: transport,
		Timeout:   timeout,
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}

	if resp, err = client.Get(link); err != nil {
		return
	}

	if resp == nil {
		err = errors.New("Response is nil. ")
		return
	}

	defer func() {
		if derr := resp.Body.Close(); derr != nil {
			if err == nil {
				err = derr
			} else {
				err = fmt.Errorf("%v %v", err, derr)
			}
		}
	}()

	statusCode = resp.StatusCode
	if statusCode != 200 {
		err = errors.New("File access error. ")
		return
	}

	b, err = ioutil.ReadAll(resp.Body)
	return
}
