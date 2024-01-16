package discfilter

import (
	"fmt"

	lru "github.com/hashicorp/golang-lru"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/enr"
)

var (
	enabled    = false
	dynamic, _ = lru.New(50000)
)

func Enable() {
	enabled = true
}

func Ban(id enode.ID) {
	fmt.Println("Banning", id)
	if enabled {
		dynamic.Add(id, struct{}{})
	}
}

func BannedDynamic(id enode.ID) bool {
	if !enabled {
		return false
	}
	return dynamic.Contains(id)
}

func BannedStatic(rec *enr.Record) bool {
	if !enabled {
		return false
	}
	return rec.Has("eth") || rec.Has("eth2")
}

func Banned(id enode.ID, rec *enr.Record) bool {
	if !enabled {
		return false
	}

	return BannedStatic(rec) || BannedDynamic(id)
}

func GetBanned() []enode.ID {
	res := make([]enode.ID, 0, dynamic.Len())
	for _, id := range dynamic.Keys() {
		res = append(res, id.(enode.ID))
	}
	return res
}
