package exampledata

type OrderDetail struct {
	ID       string   `json:"id" yaml:"id"`
	UserName string   `json:"user_name" yaml:"user_name"`
	OrderID  string   `json:"order_id" yaml:"order_id"`
	Callback string   `json:"callback" yaml:"callback"`
	Address  []string `json:"address" yaml:"address"`
}
