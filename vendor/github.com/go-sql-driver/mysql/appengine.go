// Go MySQL Driver - A MySQL-Driver for Go's database/sql package
//
// Copyright 2013 The Go-MySQL-Driver Authors. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at http://mozilla.org/MPL/2.0/.

// +build appengine

package mysql

import (
<<<<<<< HEAD
	"google.golang.org/appengine/cloudsql"
=======
	"appengine/cloudsql"
>>>>>>> 258d5c409a01370dfe542ceadc3d1669659150fe
)

func init() {
	RegisterDial("cloudsql", cloudsql.Dial)
}
