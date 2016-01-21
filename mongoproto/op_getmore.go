package mongoproto

import (
	"fmt"
	mgo "github.com/10gen/llmgo"
	"github.com/10gen/llmgo/bson"
	"io"
	"time"
)

// OpGetMore is used to query the database for documents in a collection.
// http://docs.mongodb.org/meta-driver/latest/legacy/mongodb-wire-protocol/#op-get-more
type GetMoreOp struct {
	Header MsgHeader
	mgo.GetMoreOp
}

func (op *GetMoreOp) OpCode() OpCode {
	return OpCodeGetMore
}

func (op *GetMoreOp) Meta() OpMetadata {
	return OpMetadata{"getmore", op.Collection, ""}
}

func (op *GetMoreOp) String() string {
	return fmt.Sprintf("GetMore ns:%v limit:%v cursorId:%v", op.Collection, op.Limit, op.CursorId)
}

func (op *GetMoreOp) FromReader(r io.Reader) error {
	var b [12]byte
	if _, err := io.ReadFull(r, b[:4]); err != nil {
		return err
	}
	name, err := readCStringFromReader(r)
	if err != nil {
		return err
	}
	op.Collection = string(name)
	if _, err := io.ReadFull(r, b[:12]); err != nil {
		return err
	}
	op.Limit = getInt32(b[:], 0)
	op.CursorId = getInt64(b[:], 4)
	return nil
}

func (op *GetMoreOp) Execute(session *mgo.Session) (*OpResult, error) {
	before := time.Now()

	// XXX don't actually use op.CursorID, but look up the translated cursor id from op.CursorID
	data, reply, err := mgo.ExecOpWithReply(session, &op.GetMoreOp)
	after := time.Now()

	result := &OpResult{reply, make([]bson.D, 0, len(data)), after.Sub(before)}

	for _, d := range data {
		dataDoc := bson.D{}
		err = bson.Unmarshal(d, &dataDoc)
		if err != nil {
			return nil, err
		}
		result.Docs = append(result.Docs, dataDoc)
	}

	return result, nil
}

func (getMoreOp1 *GetMoreOp) Equals(otherOp Op) bool {
	getMoreOp2, ok := otherOp.(*GetMoreOp)
	if !ok {
		return false
	}
	switch {
	case getMoreOp1.Collection != getMoreOp2.Collection:
		return false
	case getMoreOp1.Limit != getMoreOp2.Limit:
		return false
	}
	//currently doesn't compare cursorID's, not totally sure what to do about that just yet
	return true
}
