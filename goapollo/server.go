package goapollo

import (
	"net/http"
	"fmt"
	log "github.com/sirupsen/logrus"
)

type HttpServer struct {

}

func NewHttpServer() *HttpServer  {
	return &HttpServer{}
}
// 启动 HTTP 服务.
func (s *HttpServer) Start(addr string) error {
	http.HandleFunc("/ping", s.Ping)

	log.Infof("http server Running on http://%s", addr)

	return http.ListenAndServe(addr, nil)
}

func (s *HttpServer) Ping(w http.ResponseWriter, r *http.Request) {


	_,err := fmt.Fprintln(w, "pong")

	if err != nil {
		defer log.Infof("%s GET 500 %s - %s", r.RequestURI, err, r.RemoteAddr)
	} else {
		defer log.Infof("%s GET 200 OK %s", r.RequestURI,r.RemoteAddr)
	}
}

