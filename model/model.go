package model

type User struct {
	BaseInt64Model
	Name        string `gorm:"unique;not null"`
	Tel         string
	Gender      int
	CreditCards []*CreditCard `gorm:"foreignKey:UserName;references:Name"` //has many重写引用
	Languages   []*Language   `gorm:"many2many:user_languages;"`           //many to many
}

func (m *User) TableName() string {
	return "user"
}

type Language struct {
	BaseInt64Model
	Name  string  `gorm:"unique;not null"`
	Users []*User `gorm:"many2many:user_languages;"`
}

func (m *Language) TableName() string {
	return "language"
}

type CreditCard struct {
	BaseInt64Model
	CardNumber string
	UserName   string
}

type Folder struct {
	BaseInt64Model
	Name   string `gorm:"unique;not null"`
	Image  Image  `gorm:"foreignkey:FolderName;references:Name"` //反向引用外键(has one)
	IsZip  bool   `gorm:"not null"`
	IsShow bool   `gorm:"not null"`
}

type Image struct {
	BaseInt64Model
	UserName    string
	FolderName  string `gorm:"not null;comment:project名称"`
	ImageName   string `gorm:"unique;not null"`
	OldSerNum   string `gorm:"not null;comment:原始的serNum号"`
	NewSerNum   string `gorm:"comment:串起来的serNum号"`
	UpdateState int    `gorm:"not null"`
}
