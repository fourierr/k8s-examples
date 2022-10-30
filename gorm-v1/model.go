package gorm_v1

import "time"

// ConfigItem 存储配置条目
type ConfigItem struct {
	BaseModel `json:",inline"`
	// AppName indicates the application
	AppName string `json:"appName" gorm:"column:app_name"`
	// GroupId indicates the group
	GroupId string `json:"groupId" gorm:"column:group_id"`
	// Name 文件配置为文件名，环境变量配置为应用名+groupId+环境变量类型
	Name string `json:"name" gorm:"column:name"`
	// Content 配置内容
	Content string `json:"content" gorm:"column:content"`
	// ConfigType 配置内容（env/file)
	ConfigType string `json:"config_type" gorm:"column:config_type"`
	// Md5 内容的Md5值
	Md5 string `json:"md5" gorm:"column:md5"`
	// ConfigId 确保唯一，存储的是实际生效资源的后缀，前缀是组件名（命名空间下组件唯一）
	ConfigId string `json:"configId" gorm:"column:config_id"`
	// Encrypt 是否加密
	Encrypt bool `json:"encrypt" gorm:"column:encrypt"`
	// MountPath 挂载路径
	MountPath string `json:"mountPath" gorm:"column:mount_path"`
	// SubPath subPath
	SubPath string `json:"subPath" gorm:"column:sub_path"`
	// Volume 卷名, 转换过来需要指定
	VolumeName string `json:"volumeName" gorm:"column:volume_name"`
	// Comment comment
	Comment string `json:"comment"`
}

// BaseModel base model
type BaseModel struct {
	// CreateTime 创建时间
	CreateTime time.Time `gorm:"column:create_time" json:"createTime,omitempty"`
	// UpdateTime 更新时间
	UpdateTime time.Time `gorm:"column:update_time" json:"updateTime,omitempty"`
	// CreateBy 创建者
	CreateBy string `gorm:"column:create_by" json:"createBy,omitempty"`
	// UpdateBy 修改
	UpdateBy string `gorm:"column:update_by" json:"updateBy,omitempty"`
	// ID id
	ID uint `gorm:"column:id" json:"id"`
}
