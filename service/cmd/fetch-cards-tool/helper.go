package main

import (
	"fmt"
	"log"
	"os"

	"github.com/liamg/tml"
)

func exitWithError(desc string, err error, exitCode int) {
	err = fmt.Errorf("%s: %w", desc, err)

	tmlErr := tml.Printf("<red><bold>Error: %s</bold></red>\n", err.Error())
	if tmlErr != nil {
		log.Printf("tml error: %s, cause: %s \n", tmlErr.Error(), err.Error())
	}
	os.Exit(exitCode)
}
