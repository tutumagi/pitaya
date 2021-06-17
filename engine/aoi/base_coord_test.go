package aoi

import (
	"testing"

	. "github.com/go-playground/assert/v2"
)

func Test_CoordFlag(t *testing.T) {
	c := newBaseCoord(nil)

	Equal(t, c._flags, nodeFlagUnknown)

	c.addFlags(nodeFlagInstalling)
	Equal(t, c.hasFlags(nodeFlagInstalling), true)
	c.removeFlags(nodeFlagInstalling)
	Equal(t, c.hasFlags(nodeFlagInstalling), false)

	c.addFlags(nodeFlagRemoved | nodeFlagRemoving)
	Equal(t, c.hasFlags(nodeFlagRemoved), true)
	Equal(t, c.hasFlags(nodeFlagRemoving), true)
	Equal(t, c.hasFlags(nodeFlagRemoved|nodeFlagRemoving), true)
	Equal(t, c.hasFlags(nodeFlagInstalling), false)

	c.removeFlags(nodeFlagRemoved)
	Equal(t, c.hasFlags(nodeFlagRemoved), false)
	Equal(t, c.hasFlags(nodeFlagRemoving), true)
	Equal(t, c.hasFlags(nodeFlagRemoved|nodeFlagRemoving), true)

	c.removeFlags(nodeFlagRemoving)
	Equal(t, c.hasFlags(nodeFlagRemoving), false)
	Equal(t, c.hasFlags(nodeFlagRemoved|nodeFlagRemoving), false)
}
