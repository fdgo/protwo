package daemon

import "fmt"

func redirectStderr() {
	fmt.Println("do not newrpc/support redirect panic on windows")
}
