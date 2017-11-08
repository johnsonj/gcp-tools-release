package mocks

/*
 * Copyright 2017 Google Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

type TelemetrySink struct {
	RecordFn func(counters map[string]int)

	RecordCounters []map[string]int
}

func (mts *TelemetrySink) Record(counters map[string]int) {
	if mts.RecordFn != nil {
		mts.RecordFn(counters)
	}

	mts.RecordCounters = append(mts.RecordCounters, counters)
}