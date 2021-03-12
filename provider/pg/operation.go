package pg

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jackc/pgx/v4"
	uuid "github.com/satori/go.uuid"
	"go.etcd.io/bbolt"
)

func RotateInstancePassword(inst *PostgresInstance) error {
	newAddressList := make(map[string]struct {
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
	})
	// copy new
	for k, v := range inst.AddressList {
		newAddressList[k] = v
	}

	// 1. check connectivity using old, current, new passwords
	for k, v := range inst.AddressList {
		host := v.Host
		port := v.Port
		// copy new
		newUserList := make(map[string]struct {
			UserName           string `json:"user_name"`
			Password           string `json:"password"`             // current password
			OldPassword        string `json:"old_password"`         // old password
			PendingNewPassword string `json:"pending_new_password"` // new password
			PasswordExpireAt   int    `json:"password_expire_at"`
			Database           string `json:"database"`
		})
		for kk, vv := range v.UserMap {
			newUserList[kk] = vv
		}
		// check and update all user
		for kk, vv := range v.UserMap {
			newVV, err := checkAndUpdateUser(host, port, vv)
			if err != nil {
				return err
			}
			newUserList[kk] = newVV
		}

		newV := v
		newV.UserMap = newUserList
		newAddressList[k] = newV
	}
	inst.AddressList = newAddressList

	// 2. update old, current, new passwords if needed
	b, err := MarshallAndEncInstance(inst)
	if err != nil {
		log.Println("[ERROR] MarshallAndEncInstance error.", err)
		return err
	}
	err = container.DoTxWithinBucket(namespace, func(bucket *bbolt.Bucket) error {
		return bucket.Put(MakeAvailableInstanceNameKey(inst.InstanceName), b)
	})
	if err != nil {
		log.Println("[ERROR] save error.", err)
		return err
	}

	// 3. update new password in storage
	for k, v := range inst.AddressList {
		host := v.Host
		port := v.Port
		// copy new
		newUserList := make(map[string]struct {
			UserName           string `json:"user_name"`
			Password           string `json:"password"`             // current password
			OldPassword        string `json:"old_password"`         // old password
			PendingNewPassword string `json:"pending_new_password"` // new password
			PasswordExpireAt   int    `json:"password_expire_at"`
			Database           string `json:"database"`
		})
		for kk, vv := range v.UserMap {
			newUserList[kk] = vv
		}
		// check and update all user
		for kk, vv := range v.UserMap {
			newVV, err := generateNewPassword(host, port, vv)
			if err != nil {
				return err
			}
			newUserList[kk] = newVV
		}

		newV := v
		newV.UserMap = newUserList
		newAddressList[k] = newV
	}
	inst.AddressList = newAddressList
	b, err = MarshallAndEncInstance(inst)
	if err != nil {
		log.Println("[ERROR] MarshallAndEncInstance error.", err)
		return err
	}
	err = container.DoTxWithinBucket(namespace, func(bucket *bbolt.Bucket) error {
		return bucket.Put(MakeAvailableInstanceNameKey(inst.InstanceName), b)
	})
	if err != nil {
		log.Println("[ERROR] save error.", err)
		return err
	}

	// 4. update db password
	err = updatePgInstPassword(inst)
	if err != nil {
		log.Println("[ERROR] updatePgInstPassword error.", err)
		return err
	}

	// 5. update old, current, new passwords
	for k, v := range inst.AddressList {
		// copy new
		newUserList := make(map[string]struct {
			UserName           string `json:"user_name"`
			Password           string `json:"password"`             // current password
			OldPassword        string `json:"old_password"`         // old password
			PendingNewPassword string `json:"pending_new_password"` // new password
			PasswordExpireAt   int    `json:"password_expire_at"`
			Database           string `json:"database"`
		})
		for kk, vv := range v.UserMap {
			newUserList[kk] = vv
		}
		// check and update all user
		for kk, vv := range v.UserMap {
			newVV := vv
			newVV.Password = newVV.PendingNewPassword
			newVV.PendingNewPassword = ""
			newUserList[kk] = newVV
		}

		newV := v
		newV.UserMap = newUserList
		newAddressList[k] = newV
	}
	inst.AddressList = newAddressList
	b, err = MarshallAndEncInstance(inst)
	if err != nil {
		log.Println("[ERROR] MarshallAndEncInstance error.", err)
		return err
	}
	err = container.DoTxWithinBucket(namespace, func(bucket *bbolt.Bucket) error {
		return bucket.Put(MakeAvailableInstanceNameKey(inst.InstanceName), b)
	})
	if err != nil {
		log.Println("[ERROR] save error.", err)
		return err
	}

	return nil
}

func updatePgInstPassword(inst *PostgresInstance) error {
	for _, v := range inst.AddressList {
		for _, vv := range v.UserMap {
			err := updatePassword(v.Host, v.Port, vv.UserName, vv.Password, vv.PendingNewPassword, vv.Database)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func updatePassword(host string, port int, name string, password string, newPassword string, database string) error {
	tctx, cfn := context.WithTimeout(context.Background(), 10*time.Second)
	defer cfn()
	conn, err := pgx.Connect(tctx, fmt.Sprint("host=", host, " port=", port, " connect_timeout=10 user=", name, " password=", password, " database=", database))
	if err != nil {
		return err
	}
	defer func() {
		cctx, ccfn := context.WithTimeout(context.Background(), 10*time.Second)
		defer ccfn()
		conn.Close(cctx)
	}()
	_, err = conn.Exec(context.Background(), "ALTER USER "+name+" WITH PASSWORD '"+newPassword+"'")
	return err
}

func generateNewPassword(host string, port int, vv struct {
	UserName           string `json:"user_name"`
	Password           string `json:"password"`
	OldPassword        string `json:"old_password"`
	PendingNewPassword string `json:"pending_new_password"`
	PasswordExpireAt   int    `json:"password_expire_at"`
	Database           string `json:"database"`
}) (struct {
	UserName           string `json:"user_name"`
	Password           string `json:"password"`             // current password
	OldPassword        string `json:"old_password"`         // old password
	PendingNewPassword string `json:"pending_new_password"` // new password
	PasswordExpireAt   int    `json:"password_expire_at"`
	Database           string `json:"database"`
}, error) {
	newVV := vv
	newVV.PendingNewPassword = strings.ReplaceAll(uuid.NewV4().String(), "-", "")
	return newVV, nil
}

func checkAndUpdateUser(host string, port int, vv struct {
	UserName           string `json:"user_name"`
	Password           string `json:"password"`
	OldPassword        string `json:"old_password"`
	PendingNewPassword string `json:"pending_new_password"`
	PasswordExpireAt   int    `json:"password_expire_at"`
	Database           string `json:"database"`
}) (struct {
	UserName           string `json:"user_name"`
	Password           string `json:"password"`             // current password
	OldPassword        string `json:"old_password"`         // old password
	PendingNewPassword string `json:"pending_new_password"` // new password
	PasswordExpireAt   int    `json:"password_expire_at"`
	Database           string `json:"database"`
}, error) {
	newVV := vv
	if err := CheckConnectivity(host, port, vv.UserName, vv.Password, vv.Database); err == nil {
		return newVV, nil
	}
	if err := CheckConnectivity(host, port, vv.UserName, vv.OldPassword, vv.Database); err == nil {
		newVV.Password = vv.OldPassword
		return newVV, nil
	}
	if err := CheckConnectivity(host, port, vv.UserName, vv.PendingNewPassword, vv.Database); err == nil {
		newVV.Password = vv.PendingNewPassword
		newVV.PendingNewPassword = ""
		return newVV, nil
	}
	return struct {
		UserName           string `json:"user_name"`
		Password           string `json:"password"`
		OldPassword        string `json:"old_password"`
		PendingNewPassword string `json:"pending_new_password"`
		PasswordExpireAt   int    `json:"password_expire_at"`
		Database           string `json:"database"`
	}{UserName: "", Password: "", OldPassword: "", PendingNewPassword: "", PasswordExpireAt: 0, Database: ""}, errors.New("check user failed.")
}

func CheckExist(id []byte) (*PostgresInstance, bool, error) {
	var result = false
	var bb []byte
	err := container.DoTxWithinBucket(namespace, func(bucket *bbolt.Bucket) error {
		b := bucket.Get(id)
		if len(b) == 0 {
			result = false
		} else {
			result = true
			bb = b
		}
		return nil
	})
	if err != nil {
		return nil, false, err
	}
	if !result {
		return nil, false, nil
	}
	inst, err := DecAndUnmarshallInstance(bb)
	return inst, result, err
}

func CheckConnectivityBefore(inst *PostgresInstance) error {
	for _, v := range inst.AddressList {
		for u, t := range v.UserMap {
			if err := CheckConnectivity(v.Host, v.Port, u, t.Password, t.Database); err != nil {
				return err
			}
		}
	}
	return nil
}

func CheckConnectivity(host string, port int, user string, password string, database string) error {
	tctx, cfn := context.WithTimeout(context.Background(), 10*time.Second)
	defer cfn()
	conn, err := pgx.Connect(tctx, fmt.Sprint("host=", host, " port=", port, " connect_timeout=10 user=", user, " password=", password, " database=", database))
	if err != nil {
		return err
	}
	defer func() {
		cctx, ccfn := context.WithTimeout(context.Background(), 10*time.Second)
		defer ccfn()
		conn.Close(cctx)
	}()
	rows, err := conn.Query(context.Background(), "select 1")
	if err != nil {
		return err
	}
	defer rows.Close()
	if rows.Next() {
		return nil
	} else {
		return errors.New("connection validation fail")
	}
}

func MakeAvailableInstanceNameKey(instanceName string) []byte {
	return append(append([]byte{}, availableInstancePrefix...), []byte(instanceName)...)
}

func MakeDeletedInstanceNameKey(instanceName string) []byte {
	return append(append([]byte{}, deletedInstancePrefix...), []byte(instanceName)...)
}
