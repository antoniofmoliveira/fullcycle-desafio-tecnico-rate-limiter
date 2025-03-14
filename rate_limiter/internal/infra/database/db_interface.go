package database


type DBInterface interface {
	Get(key string) (string, error)
	Set(key string, value string) error

}