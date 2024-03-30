package adlist

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddNodeHead(t *testing.T) {
	l := Create()
	l.AddNodeHead("hello")
	ln := l.Index(0)
	assert.NotNil(t, ln)
	assert.Equal(t, "hello", ln.NodeValue().(string))
}
