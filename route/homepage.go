package route

import (
	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
)

func Homepage() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		var (
			b   []byte
			err error
		)
		w.WriteHeader(http.StatusOK)

		if b, err = ioutil.ReadFile("openapi.html"); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.WithFields(log.Fields{
				"route":    "Homepage",
				"locality": "ioutil.ReadFile",
			}).Error(err)
			return
		}

		if _, err = w.Write(b); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.WithFields(log.Fields{
				"route":    "Homepage",
				"locality": "ResponseWriter.Write",
			}).Error(err)
			return
		}
	}
}
