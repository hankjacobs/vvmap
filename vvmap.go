/*Package vvmap is an implementation of a delta-based CRDT map as written about in
"∆-CRDTs: Making δ-CRDTs Delta-Based" (http://nova-lincs.di.fct.unl.pt/system/publication_files/files/000/000/666/original/a12-van_der_linde.pdf?1483708753)
and "Dotted Version Vectors: Logical Clocks for Optimistic Replication" (https://arxiv.org/pdf/1011.5808.pdf).
*/
package vvmap

// ID of node. An ID must be unique across all nodes sharing the map.
type ID string

// VersionVector vector of node versions keyed by their ID
type VersionVector map[ID]uint64

// VVDot is a version vector dot used to represent an event
// with the specified version from a node with SourceID
type VVDot struct {
	SourceID ID
	Version  uint64
}

// ChooseLeftConflictResolver is a function which returns
// whether the left Record should be used to resolve
// the conflict. It can be assumed that left
// and right have the same key. This must deterministically choose the
// same item no matter the order.
type ChooseLeftConflictResolver func(key string, left, right Record) bool

// Record is a record stored in a VVMap
type Record struct {
	Key   string
	Value interface{}
	Dot   VVDot
}

// Delta are the most recent records seen between since and current
type Delta struct {
	since   VersionVector
	current VersionVector
	records []Record
}

// Map is a delta based CRDT map
type Map struct {
	resolver ChooseLeftConflictResolver
	storage  map[string]Record
	version  VersionVector
	me       ID
}

// New returns a new Map with the specified ID and conflict resolver
func New(id ID, resolver ChooseLeftConflictResolver) *Map {
	return &Map{resolver: resolver, storage: make(map[string]Record), version: make(VersionVector), me: id}
}

// Get returns the value set for key or nil if it does not exist
func (v *Map) Get(key string) interface{} {
	if record, ok := v.storage[key]; ok {
		return record.Value
	}

	return nil
}

// Set sets a value for a key
func (v *Map) Set(key string, value interface{}) {
	v.version[v.me]++
	record := Record{Key: key, Value: value, Dot: VVDot{SourceID: v.me, Version: v.version[v.me]}}
	v.storage[key] = record
}

// Keys returns all the keys contained in the map
func (v *Map) Keys() []string {
	keys := []string{}
	for key := range v.storage {
		keys = append(keys, key)
	}

	return keys
}

// Version is the map's version
func (v *Map) Version() VersionVector {
	dup := make(VersionVector)
	for k, v := range v.version {
		dup[k] = v
	}

	return dup
}

// ID is the map's ID
func (v *Map) ID() ID {
	return v.me
}

// Delta generates a list of records that have not been seen by the
// specified version vector.
func (v *Map) Delta(since VersionVector) Delta {
	records := []Record{}
	for _, record := range v.storage {
		if since[record.Dot.SourceID] < record.Dot.Version {
			records = append(records, record)
		}
	}

	return Delta{since: since, current: v.version, records: records}
}

// Merge merges a delta into the map
func (v *Map) Merge(delta Delta) {
	for _, record := range delta.records {

		if record.Dot.Version < v.version[record.Dot.SourceID] {
			continue
		}

		local, exists := v.storage[record.Key]
		if !exists || local.Dot.Version <= delta.current[local.Dot.SourceID] {
			v.storage[record.Key] = record
			continue
		}

		if v.resolver(local.Key, local, record) {
			v.storage[record.Key] = local
		} else {
			v.storage[record.Key] = record
		}
	}

	for id, version := range delta.current {
		if v.version[id] < version {
			v.version[id] = version
		}
	}
}
