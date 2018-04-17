package mongo

import (
	log "github.com/Sirupsen/logrus"
	"github.com/mulanfaas/utils"
	"github.com/mulansoft/mgodb"
	"os"
	"time"
)

func Init() {
	log.Debug("init mongo")

	addr := "mongodb://127.0.0.1:27017/mulanfaas"
	if env := os.Getenv("MONGODB"); env != "" {
		addr = env
	}

	concurrent := 128
	if env := os.Getenv("MONGODB-CONCURRENT"); env != "" {
		concurrent = utils.StringToInt(env)
	}

	timeout := 3 * time.Second
	if env := os.Getenv("MONGODB-TIMEOUT"); env != "" {
		parsed, err := time.ParseDuration(env)
		if err != nil {
			log.WithFields(log.Fields{
				"env": env,
				"err": err,
			}).Error("initMongo ParseDuration error")
		} else {
			timeout = parsed
		}
	}

	mgodb.Init(addr, concurrent, timeout)
}
