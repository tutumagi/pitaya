package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.Handle("/web/", http.StripPrefix("/web/", http.FileServer(http.Dir("web"))))

	err := http.ListenAndServe(":3251", nil)
	if err != nil {
		fmt.Printf("listen web error : %s", err)
	}
}
