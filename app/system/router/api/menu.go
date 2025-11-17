package api

import (
	"github.com/gin-gonic/gin"
	"overlink.top/app/system/conf"
	"overlink.top/app/system/model"
	"overlink.top/app/system/msg"
)

func GetMenu(c *gin.Context) {
	user := c.MustGet("identity").(*model.User)
	if user.IsSuper() {
		msg.Response(c, conf.AdminMenuList)
	} else {
		msg.Response(c, conf.CommonMenuList)
	}
}
