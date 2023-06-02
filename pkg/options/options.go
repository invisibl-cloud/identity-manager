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
	TagPrefix  string
	AWS        *awsx.Options
}

// NewOptions creates new Options
func NewOptions() *Options {
	return &Options{Tags: map[string]string{}, AWS: &awsx.Options{}}
}

// BindFlags will parse the given flagset for reconciler flags.
func (o *Options) BindFlags(fs *flag.FlagSet) {
	flag.StringVar(&o.NamePrefix, "name-prefix", "", "The resource name prefix.")
	flag.StringVar(&o.TagPrefix, "tag-prefix", "", "The resource tag prefix. note: this will be applied only to spec.tags")
	flag.Var(&o.Tags, "tag", "The resource tags. format: key=value")
	o.AWS.BindFlags(fs)
}
