// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package rtp

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSequencerBasic(t *testing.T) {
	sequencer := NewFixedSequencer(1)
	assert.Equal(t, uint16(1), sequencer.NextSequenceNumber())
	assert.Equal(t, uint64(0), sequencer.RollOverCount())
}

func TestSequencerWrapAround(t *testing.T) {
	sequencer := NewFixedSequencer(65535)
	assert.Equal(t, uint16(65535), sequencer.NextSequenceNumber())
	assert.Equal(t, uint16(0), sequencer.NextSequenceNumber())
	assert.Equal(t, uint64(1), sequencer.RollOverCount())
	assert.Equal(t, uint16(1), sequencer.NextSequenceNumber())
}

func TestSequencerMultipleRollovers(t *testing.T) {
	sequencer := NewFixedSequencer(65535)
	sequencer.NextSequenceNumber()
	sequencer.NextSequenceNumber()
	assert.Equal(t, uint64(1), sequencer.RollOverCount())

	for i := 0; i < 65536; i++ {
		sequencer.NextSequenceNumber()
	}

	assert.Equal(t, uint64(2), sequencer.RollOverCount())
}

func TestRandomSequencer(t *testing.T) {
	sequencer1 := NewRandomSequencer()
	sequencer2 := NewRandomSequencer()
	seq1 := sequencer1.NextSequenceNumber()
	seq2 := sequencer2.NextSequenceNumber()
	assert.Less(t, seq1, uint16(maxInitialRandomSequenceNumber))
	assert.Less(t, seq2, uint16(maxInitialRandomSequenceNumber))
}

func TestSequencerConcurrent(t *testing.T) {
	sequencer := NewFixedSequencer(1)
	const numGoroutines = 100
	const numIterations = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	results := make([][]uint16, numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			results[idx] = make([]uint16, numIterations)
			for j := 0; j < numIterations; j++ {
				results[idx][j] = sequencer.NextSequenceNumber()
			}
		}(i)
	}

	wg.Wait()

	seen := make(map[uint16]int)
	for i := 0; i < numGoroutines; i++ {
		for j := 0; j < numIterations; j++ {
			seen[results[i][j]]++
		}
	}

	for seq, count := range seen {
		assert.Equal(t, 1, count, "Sequence number %d appeared %d times", seq, count)

		if count != 1 {
			break
		}
	}

	assert.Equal(t, numGoroutines*numIterations, len(seen))
}

func TestSequencerRollOverCountDuringWrap(t *testing.T) {
	sequencer := NewFixedSequencer(65535)
	sequencer.NextSequenceNumber()
	assert.Equal(t, uint64(0), sequencer.RollOverCount())
	sequencer.NextSequenceNumber()
	assert.Equal(t, uint64(1), sequencer.RollOverCount())
}
