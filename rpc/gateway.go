package rpc

import (
	"log"
	"net/http"
	"reflect"
	"strconv"

	"github.com/gin-gonic/gin"
)

type SessionFunc func(req *http.Request) *Session

func DefaultSessionFunc(req *http.Request) *Session {

	uid, _ := strconv.ParseUint(req.Header.Get("token"), 10, 64)

	return &Session{Uid: uid}
}

func (s *Server) GatewayHandler(router *gin.Engine, sessionFunc SessionFunc) {

	for sname, service := range s.serviceMap {

		for mname, method := range service.method {

			router.POST(sname+"/"+mname, func() gin.HandlerFunc {

				function := method.Func
				rcvr := service.rcvr

				return func(c *gin.Context) {

					session := sessionFunc(c.Request)

					msg, err := c.GetRawData()
					if err != nil {

						log.Printf("rpcserver(%s) request body(%s) err(%v)", s.name, sname+"/"+mname, err)
						_ = c.AbortWithError(http.StatusResetContent, err)

						return
					}

					rvs := []reflect.Value{rcvr, reflect.ValueOf(msg), reflect.ValueOf(session)}
					ret := function.Call(rvs)
					resp := ret[0].Bytes()

					c.Data(http.StatusOK, c.ContentType(), resp)
				}
			}())
		}
	}
}
