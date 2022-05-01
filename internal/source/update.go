package source

import "flag"

// Update is set by the -update flag. It indicates the user running the tests
// would like to update any golden values.
var Update bool

func init() {
	flag.BoolVar(&Update, "update", false, "update golden files")
}
