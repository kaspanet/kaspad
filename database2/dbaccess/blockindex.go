package dbaccess

import "github.com/kaspanet/kaspad/database2/dbaccess/dbmodel"

func StoreBlockIndexBlock(context Context, blockIndexBlock *dbmodel.DBBlockIndexBlock) error {
	db, err := context.db()
	if err != nil {
		return err
	}
	serialized := serializeBlockIndexBlock(blockIndexBlock)
	return db.Put("kaka", serialized)
}

func serializeBlockIndexBlock(blockIndexBlock *dbmodel.DBBlockIndexBlock) []byte {
	return nil
}
