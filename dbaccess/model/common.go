package model

import "encoding/binary"

// byteOrder is the preferred byte order used for serializing numeric
// fields for storage in the database.
var byteOrder = binary.LittleEndian
