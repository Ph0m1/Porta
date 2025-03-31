package gateway

import "flag"

var config = flag.String("config", "", "input config file like ./conf/dev/")

func main() {
	flag.Parse()

}
