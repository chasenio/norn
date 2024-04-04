package main

import (
	"fmt"
	"github.com/kentio/norn/cmd"
	"os"
)

func main() {
	if err := cmd.NewApp().Run(os.Args); err != nil {
		fmt.Println(err.Error())
	}
}
