package utils

import (
	"fmt"
	"node/global"
)

func HandlePanic() {
	if r := recover(); r != nil {
		global.NODE_LOG.Error(fmt.Sprint(r))
		return
	}
}
