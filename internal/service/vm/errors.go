package vm

import "errors"

var ErrDatanodeNotOwned = errors.New("datanode not found or not owned by user")
