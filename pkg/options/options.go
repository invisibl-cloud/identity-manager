package options

import (
	"flag"

	"github.com/invisibl-cloud/identity-manager/pkg/flagx"
	"github.com/invisibl-cloud/identity-manager/pkg/providers/awsx"
)

// Options of manager
type Options struct {
	Tags       flagx.MapFlag
	NamePrefix string
	AWS        *awsx.Options
}

// NewOptions creates new Options
func NewOptions() *Options {
	return &Options{AWS: &awsx.Options{}}
}

// BindFlags will parse the given flagset for reconciler flags.
func (o *Options) BindFlags(fs *flag.FlagSet) {
	flag.Var(&o.Tags, "tags", "The resource tags.")
	flag.StringVar(&o.NamePrefix, "name-prefix", "", "The resource name prefix.")
	o.AWS.BindFlags(fs)
}
