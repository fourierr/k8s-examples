package gorm_v1

import (
	"github.com/jinzhu/gorm"
	"github.com/oam-dev/kubevela/pkg/oam/util"
	"time"
)

func main() {
	dbManager := InitMysqlDbConn()
	transactionErr := dbManager.TransactionHandle(func(db *gorm.DB) error {
		if err := dbManager.IngressClusterDao(db).CreateOrUpdate(&model2.IngressClusters{
			BaseModel: BaseModel{
				CreateTime: time.Now(),
				UpdateTime: time.Now(),
			},
			Clusters: string(util.Object2RawExtension(canaryDeployClusters).Raw),
			Flag:     true,
		}); err != nil {
			return err
		}
		return nil
	})
	if transactionErr != nil {
		panic(transactionErr)
	}
}
