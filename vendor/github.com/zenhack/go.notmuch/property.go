package notmuch

// Copyright © 2015 The go.notmuch Authors. Authors can be found in the AUTHORS file.
// Licensed under the GPLv3 or later.
// See COPYING at the root of the repository for details.

// #cgo LDFLAGS: -lnotmuch
// #include <stdlib.h>
// #include <notmuch.h>
import "C"

// Property represents a property in the database.
type Property struct {
	Key        string
	Value      string
	properties *Properties
}

func (p *Property) String() string {
	return p.Key + "=" + p.Value
}
