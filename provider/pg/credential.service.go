package pg

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PasswordResponse struct {
	AddressList map[string]struct {
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
}

func GetCredentialById(ctx *gin.Context) {
	instId := ctx.Param("id")

	inst, exist, err := CheckExist(MakeAvailableInstanceNameKey(instId))
	if err != nil {
		log.Println("[ERROR] CheckExist error.", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  1,
			"message": "internal error",
		})
		return
	}
	if !exist {
		log.Println("[ERROR] instance not exists.", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  1,
			"message": "instance not exists",
		})
		return
	}

	pr := new(PasswordResponse)
	pr.AddressList = inst.AddressList

	ctx.JSON(http.StatusOK, gin.H{
		"status": 0,
		"result": pr,
	})
}

func RotateCredentialById(ctx *gin.Context) {
	instId := ctx.Param("id")

	inst, exist, err := CheckExist(MakeAvailableInstanceNameKey(instId))
	if err != nil {
		log.Println("[ERROR] CheckExist error.", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  1,
			"message": "internal error",
		})
		return
	}
	if !exist {
		log.Println("[ERROR] instance not exists.", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  1,
			"message": "instance not exists",
		})
		return
	}

	err = RotateInstancePassword(inst)
	if err != nil {
		log.Println("[ERROR] RotateInstancePassword error.", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  1,
			"message": "internal error",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  0,
		"message": "success",
	})
}
