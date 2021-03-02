package controller

import (
	"net/http"

	"goimport.moetang.info/nekoq-security/config"

	scaffold "github.com/moetang/webapp-scaffold"

	"github.com/gin-gonic/gin"
)

func Init(scaffold *scaffold.WebappScaffold, c *config.NekoQSecurityConfig) {
	// init master key
	scaffold.GetGin().GET("/masterkey/unlock", func(ctx *gin.Context) {
		key := ctx.Query("key")

		if len(key) == 0 {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"status":  1,
				"message": "key is empty",
			})
			return
		}

		b := c.FeedShamirKey(key)
		if b {
			ctx.JSON(http.StatusOK, gin.H{
				"status":  0,
				"message": "nekoq-security is unlocked",
			})
			return
		} else {
			ctx.JSON(http.StatusOK, gin.H{
				"status":  1,
				"message": "nekoq-security is not unlocked yet",
			})
			return
		}
	})
	scaffold.GetGin().GET("/masterkey/reset_init", func(ctx *gin.Context) {
		if c.IsMasterUnlock() {
			ctx.JSON(http.StatusOK, gin.H{
				"status":  0,
				"message": "already unlocked",
			})
			return
		}

		c.ResetMasterKeyWhileUnlocking()
		ctx.JSON(http.StatusOK, gin.H{
			"status":  0,
			"message": "done",
		})
	})
}
