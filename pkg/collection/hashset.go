/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package collection

import (
	"fmt"
	"strings"
	"sync"
)

var itemExists = struct{}{}

type HashSet[T comparable] struct {
	Items map[T]struct{}
	mu    sync.RWMutex
}

func NewSet[T comparable](values ...T) *HashSet[T] {
	set := &HashSet[T]{Items: make(map[T]struct{})}
	if len(values) > 0 {
		set.Add(values...)
	}
	return set
}

func (set *HashSet[T]) Add(items ...T) {
	set.mu.Lock()
	defer set.mu.Unlock()
	for _, item := range items {
		set.Items[item] = itemExists
	}
}

func (set *HashSet[T]) Remove(items ...T) {
	set.mu.Lock()
	defer set.mu.Unlock()
	for _, item := range items {
		delete(set.Items, item)
	}
}

func (set *HashSet[T]) Contains(items ...T) bool {
	set.mu.RLock()
	defer set.mu.RUnlock()
	for _, item := range items {
		if _, contains := set.Items[item]; !contains {
			return false
		}
	}
	return true
}

func (set *HashSet[T]) Empty() bool {
	return set.Size() == 0
}

func (set *HashSet[T]) Size() int {
	return len(set.Items)
}

func (set *HashSet[T]) Clear() {
	set.Items = make(map[T]struct{})
}

func (set *HashSet[T]) Values() []T {
	values := make([]T, set.Size())
	count := 0
	for item := range set.Items {
		values[count] = item
		count++
	}
	return values
}

func (set *HashSet[T]) String() string {
	var builder strings.Builder
	builder.WriteString("HashSet\n")
	items := make([]string, 0, len(set.Items))
	for k := range set.Items {
		items = append(items, fmt.Sprintf("%v", k))
	}
	builder.WriteString(strings.Join(items, ", "))
	return builder.String()
}
