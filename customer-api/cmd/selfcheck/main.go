package main

import (
	"os"

	"smswhiteinfo/customer-api/internal/selfcheck"
)

func main() {
	os.Exit(selfcheck.Run())
}
