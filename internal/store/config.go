package store

import (
	"fmt"

	"github.com/moonmoon1919/go-api-reference/internal/config"
)

type Config struct {
	Host     config.Configurator
	User     config.Configurator
	Password config.Configurator
	Database config.Configurator
	Schema   config.Configurator
}

func (d Config) ConnectionString() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s/%s?search_path=%s",
		d.User.Must(),
		d.Password.Must(),
		d.Host.Must(),
		d.Database.Must(),
		d.Schema.Must(),
	)
}
