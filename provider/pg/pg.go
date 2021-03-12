package pg

import (
	"net/http"

	"goimport.moetang.info/nekoq-security/config"

	scaffold "github.com/moetang/webapp-scaffold"

	"github.com/gin-gonic/gin"
)

const (
	moduleName = "postgres"
	namespace  = "database.postgres"
)

var container *config.NekoQSecurityContainer

type pgModuleType struct {
}

func (p pgModuleType) SetupConfig(c *config.NekoQSecurityContainer) error {
	container = c
	return nil
}

func (p pgModuleType) InitWebScaffold(scaffold *scaffold.WebappScaffold) error {
	g := scaffold.GetGin().Group("/module/database/postgres")
	// get instances
	g.GET("/instances", wrapUnlock(ListAllInstances))
	// create instance
	g.POST("/instance", wrapUnlock(CreateInstance))
	// get instance
	g.GET("/instance/:id", wrapUnlock(GetInstanceById))
	// delete instance
	g.DELETE("/instance/:id", wrapUnlock(DeleteInstanceById))
	// update instance, without replacing existing address
	g.PUT("/instance/:id", wrapUnlock(UpdateInstanceById))

	// 1. get credential
	g.GET("/instance_credential/view/:id", wrapUnlock(GetCredentialById))
	// 2. rotate credential
	g.POST("/instance_credential/rotate/:id", wrapUnlock(RotateCredentialById))

	return nil
}

func wrapUnlock(fn func(ctx *gin.Context)) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		if !container.MasterUnlock {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"status":  -1,
				"message": "nekoq-security is not unlocked",
			})
		}
		fn(ctx)
	}
}

var pgModule config.Module = pgModuleType{}

func init() {
	config.RegisterModuleNamespace(moduleName, namespace, pgModule)
}

type PostgresInstance struct {
	InstanceName string `json:"instance_name"`
	Description  string `json:"description"`
	AddressList  map[string]struct {
		HostName string `json:"host_name"`
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Role     string `json:"role"`
		UserMap  map[string]struct {
			UserName           string `json:"user_name"`
			Password           string `json:"password"`             // current password
			OldPassword        string `json:"old_password"`         // old password
			PendingNewPassword string `json:"pending_new_password"` // new password
			PasswordExpireAt   int    `json:"password_expire_at"`
			Database           string `json:"database"`
		} `json:"user_map"`
	} `json:"address_list"`
	Status struct {
		OnlineHealth string `json:"online_health"`
	} `json:"status"`
	InstanceCredentialPolicy struct {
		Type string `json:"type"` // e.g. Default, OneTimeUser
	} `json:"instance_credential_policy"`
}
