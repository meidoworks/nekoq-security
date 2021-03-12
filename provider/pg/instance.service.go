package pg

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	"goimport.moetang.info/nekoq-security/alg/aesutils"

	"github.com/gin-gonic/gin"
	"go.etcd.io/bbolt"
)

var (
	availableInstancePrefix = []byte("pg.instance.")
	deletedInstancePrefix   = []byte("deleted.pg.instance.")
)

// list all instances
func ListAllInstances(ctx *gin.Context) {
	var r = make(map[string][]byte)
	err := container.DoTxWithinBucket(namespace, func(bucket *bbolt.Bucket) error {
		cursor := bucket.Cursor()
		for k, v := cursor.Seek(availableInstancePrefix); k != nil && bytes.HasPrefix(k, availableInstancePrefix); k, v = cursor.Next() {
			r[string(k)] = v
		}
		return nil
	})
	if err != nil {
		log.Println("[ERROR] List all instances error.", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  1,
			"message": "list all instances error",
		})
		return
	}
	var result = make(map[string]*PostgresInstance)
	for k, v := range r {
		inst, err := DecAndUnmarshallInstance(v)
		if err != nil {
			log.Println("[ERROR] decrypt instance error.", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"status":  1,
				"message": "list all instances error",
			})
			return
		}
		desensitization(inst)
		result[k] = inst
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": 0,
		"result": result,
	})
}

// create a new instance
// content-type: json
func CreateInstance(ctx *gin.Context) {
	inst := new(PostgresInstance)
	if err := ctx.ShouldBindJSON(inst); err != nil {
		log.Println("[ERROR] bind json error.", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  1,
			"message": "parameter error",
		})
		return
	}

	//check parameter
	checkInstParameter(inst)

	_, exist, err := CheckExist(MakeAvailableInstanceNameKey(inst.InstanceName))
	if err != nil {
		log.Println("[ERROR] CheckExist error.", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  1,
			"message": "internal error",
		})
		return
	}
	if exist {
		log.Println("[ERROR] instance exists.", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  1,
			"message": "instance exists",
		})
		return
	}

	if err := CheckConnectivityBefore(inst); err != nil {
		log.Println("[ERROR] CheckConnectivityBefore error.", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  1,
			"message": "cannot connect database error",
		})
		return
	}

	b, err := MarshallAndEncInstance(inst)
	if err != nil {
		log.Println("[ERROR] MarshallAndEncInstance error.", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  1,
			"message": "internal error",
		})
		return
	}

	err = container.DoTxWithinBucket(namespace, func(bucket *bbolt.Bucket) error {
		return bucket.Put(MakeAvailableInstanceNameKey(inst.InstanceName), b)
	})
	if err != nil {
		log.Println("[ERROR] save error.", err)
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

func checkInstParameter(inst *PostgresInstance) {
}

func checkInstParameterWithoutCredential(inst *PostgresInstance) {
}

// get instance by id
func GetInstanceById(ctx *gin.Context) {
	instId := ctx.Param("id")
	var r []byte
	err := container.DoTxWithinBucket(namespace, func(bucket *bbolt.Bucket) error {
		oldKey := MakeAvailableInstanceNameKey(instId)
		v := bucket.Get(oldKey)
		if v != nil {
			r = v
		}
		return nil
	})
	if err != nil {
		log.Println("[ERROR] get instance error.", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  1,
			"message": "get instance error",
		})
		return
	}

	if len(r) == 0 {
		log.Println("[ERROR] get instance error.", err)
		ctx.JSON(http.StatusNotFound, gin.H{
			"status":  1,
			"message": "not found",
		})
		return
	}

	inst, err := DecAndUnmarshallInstance(r)
	if err != nil {
		log.Println("[ERROR] DecAndUnmarshallInstance error.", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  1,
			"message": "get instance error",
		})
		return
	}

	//Desensitization
	{
		desensitization(inst)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": 0,
		"result": inst,
	})
}

func desensitization(inst *PostgresInstance) {
	addrList := inst.AddressList
	for i, v := range addrList {
		um := v.UserMap
		newUm := make(map[string]struct {
			UserName           string `json:"user_name"`
			Password           string `json:"password"`
			OldPassword        string `json:"old_password"`
			PendingNewPassword string `json:"pending_new_password"`
			PasswordExpireAt   int    `json:"password_expire_at"`
			Database           string `json:"database"`
		})
		for k, v := range um {
			newUm[k] = struct {
				UserName           string `json:"user_name"`
				Password           string `json:"password"`
				OldPassword        string `json:"old_password"`
				PendingNewPassword string `json:"pending_new_password"`
				PasswordExpireAt   int    `json:"password_expire_at"`
				Database           string `json:"database"`
			}{UserName: v.UserName, Password: "", OldPassword: "", PendingNewPassword: "", PasswordExpireAt: v.PasswordExpireAt, Database: v.Database}
		}
		v.UserMap = newUm
		addrList[i] = v
	}
	inst.AddressList = addrList
}

// delete an instance
func DeleteInstanceById(ctx *gin.Context) {
	instId := ctx.Param("id")
	err := container.DoTxWithinBucket(namespace, func(bucket *bbolt.Bucket) error {
		oldKey := MakeAvailableInstanceNameKey(instId)
		v := bucket.Get(oldKey)
		if v != nil {
			err := bucket.Delete(oldKey)
			if err != nil {
				return err
			}
			err = bucket.Put(MakeDeletedInstanceNameKey(instId), v)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		log.Println("[ERROR] delete instance error.", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  1,
			"message": "delete instance error",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  0,
		"message": "success",
	})
}

// update instance by id
func UpdateInstanceById(ctx *gin.Context) {
	instId := ctx.Param("id")

	inst := new(PostgresInstance)
	if err := ctx.ShouldBindJSON(inst); err != nil {
		log.Println("[ERROR] bind json error.", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  1,
			"message": "parameter error",
		})
		return
	}
	//check parameter
	checkInstParameterWithoutCredential(inst)
	inst.InstanceName = instId

	origInst, exist, err := CheckExist(MakeAvailableInstanceNameKey(instId))
	if err != nil {
		log.Println("[ERROR] CheckExist error.", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  1,
			"message": "internal error",
		})
		return
	}
	if !exist {
		log.Println("[ERROR] not exist error.", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  1,
			"message": "instance does not exist",
		})
		return
	}

	// update
	for k, v := range origInst.AddressList {
		_, ok := inst.AddressList[k]
		if ok {
			inst.AddressList[k] = v
		}
	}

	if err := CheckConnectivityBefore(inst); err != nil {
		log.Println("[ERROR] CheckConnectivityBefore error.", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  1,
			"message": "cannot connect database error",
		})
		return
	}

	b, err := MarshallAndEncInstance(inst)
	if err != nil {
		log.Println("[ERROR] MarshallAndEncInstance error.", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  1,
			"message": "internal error",
		})
		return
	}

	err = container.DoTxWithinBucket(namespace, func(bucket *bbolt.Bucket) error {
		return bucket.Put(MakeAvailableInstanceNameKey(inst.InstanceName), b)
	})
	if err != nil {
		log.Println("[ERROR] save error.", err)
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

func MarshallAndEncInstance(instance *PostgresInstance) ([]byte, error) {
	b, err := json.Marshal(instance)
	if err != nil {
		return nil, err
	}
	encb, err := aesutils.Encrypt(b, container.MasterKey)
	if err != nil {
		return nil, err
	}
	return encb, nil
}

func DecAndUnmarshallInstance(b []byte) (*PostgresInstance, error) {
	decb, err := aesutils.Decrypt(b, container.MasterKey)
	if err != nil {
		return nil, err
	}
	inst := new(PostgresInstance)
	err = json.Unmarshal(decb, inst)
	if err != nil {
		return nil, err
	}
	return inst, nil
}
