package base

type Base_in struct {
	Name string `json:"name,omitempty" gorm:"index" xlsx:"#"`
}
type Base_out struct {
	Name string `json:"name,omitempty" gorm:"index" xlsx:"#"`
}
