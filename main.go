package main

import (
    //"fmt"
    "os"

    "gh-migrate-variables/cmd"
)

func main() {
    if err := cmd.Execute(); err != nil {
        os.Exit(1)
    }
}

