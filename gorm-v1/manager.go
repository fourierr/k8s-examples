package gorm_v1

import (
	"fmt"
	"k8s.io/klog/v2"
	"log"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

type Manager struct {
	db      *gorm.DB
	initOne sync.Once
	models  []model.Interface
}

// InitMysqlDbConn new mysql db
func InitMysqlDbConn() *Manager {
	var db *gorm.DB
	user := "root"
	pwd := "123456"
	host := "99.15.138.55"
	port := "3306"
	dbName := "dbName"

	DSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Asia%%2FShanghai",
		user, pwd, host, port, dbName)
	db, err := gorm.Open("mysql", DSN)
	if err != nil {
		log.Fatal(err.Error())
	}
	// 链接过期的太过频繁，会有性能问题
	db.DB().SetConnMaxLifetime(time.Duration(240) * time.Second)
	db.DB().SetConnMaxIdleTime(time.Duration(240) * time.Second)
	db.DB().SetMaxOpenConns(80)
	db.DB().SetMaxIdleConns(20)
	db = db.Debug()
	manager := &Manager{
		db:      db,
		initOne: sync.Once{},
	}
	manager.RegisterTableModel()
	return manager
}

func (m *Manager) CloseManager() error {
	return m.db.Close()
}

func (m *Manager) RegisterTableModel() {
	// Config
	m.models = append(m.models, &ConfigItem{})
}

func (m *Manager) Begin() *gorm.DB {
	return m.db.Begin()
}

func (m *Manager) DB() *gorm.DB {
	return m.db
}

func (m *Manager) EnsureEndTransactionFunc() func(tx *gorm.DB) {
	return func(tx *gorm.DB) {
		if r := recover(); r != nil {
			klog.Errorf("Unexpected panic occurred, rollback transaction: %v", r)
			tx.Rollback()
		}
	}
}

func (m *Manager) TransactionHandle(callback func(db *gorm.DB) error) (err error) {
	tx := m.Begin()
	defer func() {
		if r := recover(); r != nil {
			klog.Errorf("Unexpected panic occurred, rollback transaction: %v", r)
			err = fmt.Errorf("%v", r)
		}
		if err != nil {
			rollBackErr := tx.Rollback().Error
			if rollBackErr == nil {
				err = fmt.Errorf("callback err : (%s) , rollBack success", err.Error())
			} else {
				err = fmt.Errorf("callback err : (%s) , rollBack err : (%s) ", err.Error(), rollBackErr.Error())
			}
		} else {
			if err = tx.Commit().Error; err != nil {
				err = fmt.Errorf("commit err : (%s)", err.Error())
			}
		}
	}()
	err = callback(tx)
	return err
}
