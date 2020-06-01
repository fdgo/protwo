package user

type User_in struct {
	Name string `json:"name,omitempty" gorm:"index" xlsx:"#"`
}
type User_out struct {
	Name string `json:"name,omitempty" gorm:"index" xlsx:"#"`
}
