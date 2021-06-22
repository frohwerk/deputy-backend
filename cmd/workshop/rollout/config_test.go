package rollout_test

import "github.com/frohwerk/deputy-backend/internal/logger"

var Log logger.Logger = logger.Noop

func init() {
	Log = logger.Basic(logger.LEVEL_INFO)
	// Log = logger.Basic(logger.LEVEL_DEBUG)
	// Log = logger.Basic(logger.LEVEL_TRACE)
}
