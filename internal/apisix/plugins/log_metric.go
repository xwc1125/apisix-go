// Package plugins
//
// @author: xwc1125
package plugins

import "github.com/chain5j/logger"

func log() logger.Logger {
	return logger.Log("plugins")
}
