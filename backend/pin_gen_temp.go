package main

import (
"fmt"
"golang.org/x/crypto/bcrypt"
)

func main() {
hashedPin, err := bcrypt.GenerateFromPassword([]byte("1234"), bcrypt.DefaultCost)
if err != nil {
panic(err)
}
fmt.Printf("%s", string(hashedPin))
}
