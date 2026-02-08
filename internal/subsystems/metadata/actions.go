package metadata

import (
	"log/slog"
	"strings"

	"github.com/dgraph-io/badger/v4"
	"gitlab.com/stexxo/dynocue/dynod/internal/bus"
)

func (m *Metadata) SetMetadataValue(msg bus.Message) {

	msgSplit := strings.Split(msg.Subject, ".")
	if len(msgSplit) < 4 {
		slog.Warn("set metadata value message received with less than 4 items in subject")
		return
	}

	metadataAttr := strings.Join(msgSplit[3:], ".")
	db, err := m.show.GetDatabase()
	if err != nil {
		slog.Error("failed to get database", "error", err)
		return
	}

	err = db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(showMetadataKey+":"+metadataAttr), msg.Data)
	})
	if err != nil {
		slog.Error("failed to set metadata value", "error", err)
		return
	}

	m.evMgr.SendHelper("show.metadata.value."+metadataAttr, msg.Data)
}

func (m *Metadata) GetMetadataValue(msg bus.Message) {
	msgSplit := strings.Split(msg.Subject, ".")
	if len(msgSplit) < 4 {
		slog.Warn("get metadata value message received with less than 4 items in subject")
		return
	}
	metadataAttr := strings.Join(msgSplit[3:], ".")

	var res string
	db, err := m.show.GetDatabase()
	if err != nil {
		slog.Error("failed to get database", "error", err)
		return
	}

	err = db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(showMetadataKey + ":" + metadataAttr))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			res = string(val)
			return nil
		})
	})
	if err != nil {
		slog.Error("failed to set metadata value", "error", err)
		return
	}
	m.evMgr.RespondHelper(msg, []byte(res))
}
