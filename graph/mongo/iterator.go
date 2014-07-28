// Copyright 2014 The Cayley Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mongo

import (
	"fmt"
	"strings"

	"github.com/barakmich/glog"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/google/cayley/graph"
	"github.com/google/cayley/graph/iterator"
	"github.com/google/cayley/quad"
)

type Iterator struct {
	iterator.Base
	qs         *TripleStore
	dir        quad.Direction
	iter       *mgo.Iter
	hash       string
	name       string
	size       int64
	isAll      bool
	constraint bson.M
	collection string
}

func NewIterator(qs *TripleStore, collection string, d quad.Direction, val graph.Value) *Iterator {
	var m Iterator
	iterator.BaseInit(&m.Base)

	m.name = qs.NameOf(val)
	m.collection = collection
	switch d {
	case quad.Subject:
		m.constraint = bson.M{"Subject": m.name}
	case quad.Predicate:
		m.constraint = bson.M{"Predicate": m.name}
	case quad.Object:
		m.constraint = bson.M{"Object": m.name}
	case quad.Label:
		m.constraint = bson.M{"Label": m.name}
	}

	m.qs = qs
	m.dir = d
	m.iter = qs.db.C(collection).Find(m.constraint).Iter()
	size, err := qs.db.C(collection).Find(m.constraint).Count()
	if err != nil {
		glog.Errorln("Trouble getting size for iterator! ", err)
		return nil
	}
	m.size = int64(size)
	m.hash = val.(string)
	m.isAll = false
	return &m
}

func NewAllIterator(qs *TripleStore, collection string) *Iterator {
	var m Iterator
	m.qs = qs
	m.dir = quad.Any
	m.constraint = nil
	m.collection = collection
	m.iter = qs.db.C(collection).Find(nil).Iter()
	size, err := qs.db.C(collection).Count()
	if err != nil {
		glog.Errorln("Trouble getting size for iterator! ", err)
		return nil
	}
	m.size = int64(size)
	m.hash = ""
	m.isAll = true
	return &m
}

func (it *Iterator) Reset() {
	it.iter.Close()
	it.iter = it.qs.db.C(it.collection).Find(it.constraint).Iter()

}

func (it *Iterator) Close() {
	it.iter.Close()
}

func (it *Iterator) Clone() graph.Iterator {
	var newM graph.Iterator
	if it.isAll {
		newM = NewAllIterator(it.qs, it.collection)
	} else {
		newM = NewIterator(it.qs, it.collection, it.dir, it.hash)
	}
	newM.CopyTagsFrom(it)
	return newM
}

func (it *Iterator) Next() (graph.Value, bool) {
	var result struct {
		Id string "_id"
		//Sub string "Sub"
		//Pred string "Pred"
		//Obj string "Obj"
	}
	found := it.iter.Next(&result)
	if !found {
		err := it.iter.Err()
		if err != nil {
			glog.Errorln("Error Nexting Iterator: ", err)
		}
		return nil, false
	}
	it.Last = result.Id
	return result.Id, true
}

func (it *Iterator) Check(v graph.Value) bool {
	graph.CheckLogIn(it, v)
	if it.isAll {
		it.Last = v
		return graph.CheckLogOut(it, v, true)
	}
	var offset int
	switch it.dir {
	case quad.Subject:
		offset = 0
	case quad.Predicate:
		offset = (it.qs.hasher.Size() * 2)
	case quad.Object:
		offset = (it.qs.hasher.Size() * 2) * 2
	case quad.Label:
		offset = (it.qs.hasher.Size() * 2) * 3
	}
	val := v.(string)[offset : it.qs.hasher.Size()*2+offset]
	if val == it.hash {
		it.Last = v
		return graph.CheckLogOut(it, v, true)
	}
	return graph.CheckLogOut(it, v, false)
}

func (it *Iterator) Size() (int64, bool) {
	return it.size, true
}

var mongoType graph.Type

func init() {
	mongoType = graph.RegisterIterator("mongo")
}

func Type() graph.Type { return mongoType }

func (it *Iterator) Type() graph.Type {
	if it.isAll {
		return graph.All
	}
	return mongoType
}

func (it *Iterator) Sorted() bool                     { return true }
func (it *Iterator) Optimize() (graph.Iterator, bool) { return it, false }

func (it *Iterator) DebugString(indent int) string {
	size, _ := it.Size()
	return fmt.Sprintf("%s(%s size:%d %s %s)", strings.Repeat(" ", indent), it.Type(), size, it.hash, it.name)
}

func (it *Iterator) Stats() graph.IteratorStats {
	size, _ := it.Size()
	return graph.IteratorStats{
		CheckCost: 1,
		NextCost:  5,
		Size:      size,
	}
}
