// Copyright 2018 Adknown Inc. All rights reserved.
// Created:  06/06/18
// Author:   matt
// Project:  gen

package scalyr

import (
	"os"
)

const (
	testHost       = "goScalyrApi"
	testTier       = "test"
	testSystem     = "goAPITesting"
	testThreadName = "Test Thread"
	testLogfile    = "/arbitary/file/path.txt"
	testParser     = "json"
)

var (
	testReadKey = os.Getenv("SCALYR_READ_KEY")
)
