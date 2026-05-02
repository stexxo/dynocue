package db

import (
	"bytes"
	"compress/gzip"
	"encoding/gob"
	"io"

	"github.com/hashicorp/go-memdb"
)

func SerializeTable(db *memdb.MemDB, tableName string) (*bytes.Buffer, error) {
	var buf bytes.Buffer

	// Create gzip writer
	gw := gzip.NewWriter(&buf)
	enc := gob.NewEncoder(gw)

	txn := db.Txn(false)
	defer txn.Abort()

	// Get an iterator for the entire table using the primary index
	// Note: gomemdb usually requires an index name. Default is often "id".
	it, err := txn.Get(tableName, "id")
	if err != nil {
		return nil, err
	}

	for obj := it.Next(); obj != nil; obj = it.Next() {
		if err := enc.Encode(obj); err != nil {
			return nil, err
		}
	}

	// Important: Close gzip writer to flush headers/trailers before reading buffer
	if err := gw.Close(); err != nil {
		return nil, err
	}

	return &buf, nil
}

func RestoreTable(db *memdb.MemDB, tableName string, data io.Reader) error {
	gr, err := gzip.NewReader(data)
	if err != nil {
		return err
	}
	defer gr.Close()

	dec := gob.NewDecoder(gr)
	txn := db.Txn(true)

	for {
		// Use a pointer to your specific struct type here
		// or a generic interface if gob.Register was used
		var obj interface{}
		err := dec.Decode(&obj)
		if err == io.EOF {
			break
		}
		if err != nil {
			txn.Abort()
			return err
		}

		if err := txn.Insert(tableName, obj); err != nil {
			txn.Abort()
			return err
		}
	}

	txn.Commit()
	return nil
}
