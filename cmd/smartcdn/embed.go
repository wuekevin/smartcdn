package main

import (
	"io/fs"

	"github.com/wuekevin/smartcdn/assets"
)

func frontendFS() fs.FS {
	return assets.Frontend()
}
