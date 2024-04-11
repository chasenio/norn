package main

import (
	"fmt"
	"github.com/kentio/norn/cmd/pick"
	"os"
)

func main() {
	if err := pick.NewApp().Run(os.Args); err != nil {
		fmt.Println(err.Error())
	}
}
