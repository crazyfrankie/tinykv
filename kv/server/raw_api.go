package server

import (
	"context"

	"github.com/pingcap-incubator/tinykv/kv/storage"
	"github.com/pingcap-incubator/tinykv/proto/pkg/kvrpcpb"
)

// The functions below are Server's Raw API. (implements TinyKvServer).
// Some helper methods can be found in sever.go in the current directory

// RawGet return the corresponding Get response based on RawGetRequest's CF and Key fields
func (server *Server) RawGet(_ context.Context, req *kvrpcpb.RawGetRequest) (*kvrpcpb.RawGetResponse, error) {
	res := &kvrpcpb.RawGetResponse{}

	reader, err := server.storage.Reader(req.Context)
	if err != nil {
		return nil, err
	}

	res.Value, err = reader.GetCF(req.GetCf(), req.GetKey())
	if err != nil {
		return res, err
	}

	res.NotFound = res.Value == nil

	return res, nil
}

// RawPut puts the target data into storage and returns the corresponding response
func (server *Server) RawPut(_ context.Context, req *kvrpcpb.RawPutRequest) (*kvrpcpb.RawPutResponse, error) {
	res := &kvrpcpb.RawPutResponse{}

	putData := storage.Put{
		Key:   req.GetKey(),
		Value: req.GetValue(),
		Cf:    req.GetCf(),
	}

	batch := []storage.Modify{
		{Data: putData},
	}

	if err := server.storage.Write(req.Context, batch); err != nil {
		return nil, err
	}

	return res, nil
}

// RawDelete delete the target data from storage and returns the corresponding response
func (server *Server) RawDelete(_ context.Context, req *kvrpcpb.RawDeleteRequest) (*kvrpcpb.RawDeleteResponse, error) {
	res := &kvrpcpb.RawDeleteResponse{}

	deleteData := storage.Delete{
		Key: req.GetKey(),
		Cf:  req.GetCf(),
	}

	batch := []storage.Modify{
		{Data: deleteData},
	}

	if err := server.storage.Write(req.Context, batch); err != nil {
		return nil, err
	}

	return res, nil
}

// RawScan scan the data starting from the start key up to limit. and return the corresponding result
func (server *Server) RawScan(_ context.Context, req *kvrpcpb.RawScanRequest) (*kvrpcpb.RawScanResponse, error) {
	res := &kvrpcpb.RawScanResponse{}

	reader, err := server.storage.Reader(req.Context)
	if err != nil {
		return nil, err
	}

	iterator := reader.IterCF(req.GetCf())
	iterator.Seek(req.GetStartKey())

	kvs := make([]*kvrpcpb.KvPair, 0, req.GetLimit())

	for i := uint32(0); iterator.Valid() && i < req.GetLimit(); i++ {
		item := iterator.Item()

		k := item.Key()
		v, err := item.Value()
		if err != nil {
			return nil, err
		}
		kvs = append(kvs, &kvrpcpb.KvPair{
			Key:   k,
			Value: v,
		})
		iterator.Next()
	}

	res.Kvs = kvs

	return res, nil
}
