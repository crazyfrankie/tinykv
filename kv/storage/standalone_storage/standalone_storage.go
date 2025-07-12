package standalone_storage

import (
	"errors"
	"path/filepath"

	"github.com/Connor1996/badger"
	"github.com/pingcap-incubator/tinykv/kv/config"
	"github.com/pingcap-incubator/tinykv/kv/storage"
	"github.com/pingcap-incubator/tinykv/kv/util/engine_util"
	"github.com/pingcap-incubator/tinykv/proto/pkg/kvrpcpb"
)

// StandAloneStorage is an implementation of `Storage` for a single-node TinyKV instance. It does not
// communicate with other nodes and all data is stored locally.
type StandAloneStorage struct {
	config *config.Config
	engine *engine_util.Engines
}

type StandAloneReader struct {
	kvTxn *badger.Txn
}

func NewStandAloneStorage(conf *config.Config) *StandAloneStorage {
	kvPath := filepath.Join(conf.DBPath, "kv")
	raftPath := filepath.Join(conf.DBPath, "raft")

	kvEngine := engine_util.CreateDB(kvPath, false)
	raftEngine := engine_util.CreateDB(raftPath, true)

	return &StandAloneStorage{
		config: conf,
		engine: engine_util.NewEngines(kvEngine, raftEngine, kvPath, raftPath),
	}
}

func NewStandAloneReader(txn *badger.Txn) *StandAloneReader {
	return &StandAloneReader{kvTxn: txn}
}

func (s *StandAloneStorage) Start() error {
	return nil
}

func (s *StandAloneStorage) Stop() error {
	return s.engine.Close()
}

func (s *StandAloneStorage) Reader(ctx *kvrpcpb.Context) (storage.StorageReader, error) {
	txn := s.engine.Kv.NewTransaction(false)
	return NewStandAloneReader(txn), nil
}

func (s *StandAloneStorage) Write(ctx *kvrpcpb.Context, batch []storage.Modify) error {
	var err error

	for _, b := range batch {
		switch b.Data.(type) {
		case storage.Put:
			err = engine_util.PutCF(s.engine.Kv, b.Cf(), b.Key(), b.Value())
		case storage.Delete:
			err = engine_util.DeleteCF(s.engine.Kv, b.Cf(), b.Key())
		}
		if err != nil {
			return err
		}
	}

	return nil
}

/*
	Implementation for StorageReader
*/

func (s *StandAloneReader) GetCF(cf string, key []byte) ([]byte, error) {
	// In here, we do not just call `return engine_util.GetCFFromTxn(s.kvTxn, cf, key)`,
	// because test for NotFound case needed receive `nil`
	value, err := engine_util.GetCFFromTxn(s.kvTxn, cf, key)
	if errors.Is(err, badger.ErrKeyNotFound) {
		return nil, nil
	}
	return value, err
}

func (s *StandAloneReader) IterCF(cf string) engine_util.DBIterator {
	return engine_util.NewCFIterator(cf, s.kvTxn)
}

func (s *StandAloneReader) Close() {
	// similar to `commit`
	s.kvTxn.Discard()
}
