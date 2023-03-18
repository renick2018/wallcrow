package model

type OpenaiAccount struct {
	BaseModel `json:",inline"`
	Email     string `json:"email" gorm:"comment: 邮箱"`
	Password  string `json:"password" gorm:"comment: 密码"`
	Apikey    string `json:"apikey" gorm:"comment: key"`
	Plus      bool   `json:"plus" gorm:"comment: plus"`
	UseUp     bool   `json:"free_out" gorm:"comment: 额度用尽"`
	Status    string `json:"status" gorm:"comment: 状态"`
}
