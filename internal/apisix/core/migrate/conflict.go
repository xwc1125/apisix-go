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

package migrate

import (
	"context"

	"github.com/chain5j/logger"
	store2 "github.com/xwc1125/apisix-go/internal/apisix/core/store"
)

func isConflicted(ctx context.Context, new *DataSet) (bool, *DataSet) {
	isConflict := false
	conflictedData := newDataSet()
	store2.RangeStore(func(key store2.HubKey, s *store2.GenericStore) bool {
		new.rangeData(key, func(i int, obj interface{}) bool {
			// Only check key of store conflict for now.
			// TODO: Maybe check name of some entiries.
			_, err := s.CreateCheck(obj)
			if err != nil {
				isConflict = true
				err = conflictedData.Add(obj)
				if err != nil {
					logger.Error("Add obj to conflict list failed", "err", err)
					return true
				}
			}
			return true
		})
		return true
	})
	return isConflict, conflictedData
}
