package tqsdk

type Resource struct {
	Size int
}

var connStr string

func (r Resource) AsConfig() Config {
	if r.Size == 1 {
		return Config{
			ConnStr: connStr,
		}
	}

	return Config{}
}

type Config struct {
	ConnStr string
}
