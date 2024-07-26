package main

import "os"

func main() {
	c := newClient()
	defer c.close()

	if len(os.Args) == 1 {
		c.interactive()
		return
	}

	c.direct()
}
