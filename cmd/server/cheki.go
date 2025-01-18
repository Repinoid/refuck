package main

import (
	"fmt"
	"gorono/internal/basis"
	"gorono/internal/memos"
	"gorono/internal/models"
)

func main() {

	var face models.Inter

	face = basis.DBstruct{}

	face = memos.MemoryStorageStruct{}

	fmt.Printf("%+v\n", face)

}
