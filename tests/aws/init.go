package tests

import (
	"gobase-lambda/log"
	"gobase-lambda/utils/testutils"
)

func init() {
	testutils.Initialize("../../config/staging.json")
	logger := log.NewLogger(false, log.DEBUG, nil)
	log.SetDefaultLogger(logger)
}
