package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"overlink.top/app/system/conf"
	"overlink.top/app/system/log"
	"overlink.top/app/system/logic"
	"overlink.top/app/system/model"
	"overlink.top/app/system/router"
)

func main() {
	conf.InitConf()
	log.InitCore(conf.AppConf.Log)

	model.InitDb(conf.AppConf.Database)
	logic.Init()
	gin.SetMode(gin.ReleaseMode)

	routerInit := router.InitRouter()

	// Set trusted proxies for Gin
	if len(conf.AppConf.Server.TrustedProxies) > 0 {
		if err := routerInit.SetTrustedProxies(conf.AppConf.Server.TrustedProxies); err != nil {
			log.StdErrorf("Failed to set trusted proxies: %v", err)
		}
	}

	endPoint := fmt.Sprintf("%s:%d", conf.AppConf.Server.Host, conf.AppConf.Server.Port)
	server := &http.Server{
		Addr:    endPoint,
		Handler: routerInit,
	}

	var err error
	if conf.AppConf.Server.Https {
		log.StdInfof("start https server listening %s", endPoint)
		sslCertFile := conf.AbsPath(conf.AppConf.Server.SSLCertPem)
		sslKeyFile := conf.AbsPath(conf.AppConf.Server.SSLKeyPem)
		err = server.ListenAndServeTLS(sslCertFile, sslKeyFile)
	} else {
		log.StdInfof("start http server listening %s", endPoint)
		err = server.ListenAndServe()
	}

	if err != nil {
		log.StdError(err)
		os.Exit(0)
	}
}
