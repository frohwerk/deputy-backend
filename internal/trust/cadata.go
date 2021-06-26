package trust

import (
	"fmt"
	"os"
	"time"
)

var locations = []string{
	"/run/secrets/kubernetes.io/serviceaccount/ca.crt",
	"E:/projects/go/src/github.com/frohwerk/deputy-backend/certificates/minishift.crt",
}

var CAData []byte

func init() {
	for _, loc := range locations {
		_, err := os.Stat(loc)
		switch {
		case err == nil:
			if data, err := os.ReadFile(loc); err != nil {
				fmt.Fprintf(os.Stderr, "error reading cafile %s: %s\n", loc, err)
				continue
			} else {
				CAData = data
				return
			}
		case os.IsNotExist(err):
			continue
		default:
			fmt.Fprintf(os.Stderr, "error checking cafile candidate %s: %s\n", loc, err)
			continue
		}
	}
	fmt.Println("No cadata found!")
	<-time.NewTimer(15 * time.Second).C
	os.Exit(1)
}
