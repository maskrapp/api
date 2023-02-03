package healthcheck

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/maskrapp/api/internal/global"
	"github.com/sirupsen/logrus"
)

func New(ctx global.Context) http.Server {
	handler := func(w http.ResponseWriter, r *http.Request) {
		var redisErr bool
		var dbErr bool

		wg := sync.WaitGroup{}

		wg.Add(1)
		go func() {
			defer wg.Done()
			c, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()
			err := ctx.Instances().Redis.Ping(c).Err()
			if err != nil {
				logrus.Errorf("redis healthcheck failed: %v", err)
				redisErr = true
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			db, err := ctx.Instances().Gorm.DB()
			if err != nil {
				logrus.Errorf("db healthcheck failed: %v", err)
				dbErr = true
				return
			}
			c, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()
			if err := db.PingContext(c); err != nil {
				logrus.Errorf("db healthcheck failed: %v", err)
				dbErr = true
			}
		}()
		wg.Wait()
		if dbErr || redisErr {
			w.WriteHeader(500)
			return
		}
		w.Write([]byte("healthy"))
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)
	return http.Server{
		Addr:    ":9000",
		Handler: mux,
	}
}
