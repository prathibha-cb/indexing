//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package indexer

type errCode int16

const (
	ERROR_PANIC errCode = iota

	//Slab Manager
	ERROR_SLAB_INIT
	ERROR_SLAB_BAD_ALLOC_REQUEST
	ERROR_SLAB_INTERNAL_ALLOC_ERROR
	ERROR_SLAB_MEM_LIMIT_EXCEED
	ERROR_SLAB_INTERNAL_ERROR
)

type errSeverity int16

const (
	FATAL errSeverity = iota
	NORMAL
)

type errCategory int16

const (
	MESSAGING errCategory = iota
	STORAGE
	MUTATION_QUEUE
	TOPOLOGY
	STREAM_SERVER
	SLAB_MANAGER
)

type Error struct {
	code     errCode
	severity errSeverity
	category errCategory
	cause    error
	msg      string
}
