package main

import (
    "fmt"
    "golang.org/x/crypto/bcrypt"
)

func main() {
    hash, err := bcrypt.GenerateFromPassword([]byte("testpass"), bcrypt.DefaultCost)
    if err != nil {
        fmt.Println("Error generating hash:", err)
        return
    }
    fmt.Println(string(hash))
}
