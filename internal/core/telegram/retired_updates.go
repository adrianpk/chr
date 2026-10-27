package telegram

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"math"
	"sort"
)

type retiredInterval [2]int64

func readRetired(ctx context.Context, tx *sql.Tx, botID int64) ([]retiredInterval, error) {
	var encoded string
	if err := tx.QueryRowContext(ctx, `SELECT retired_update_ranges FROM telegram_receivers WHERE bot_id=? AND schema_version=1`, botID).Scan(&encoded); err != nil {
		return nil, err
	}

	if len(encoded) > 65536 {
		return nil, errors.New("telegram retired identity metadata exceeds its bound")
	}

	var pairs [][]int64
	if err := json.Unmarshal([]byte(encoded), &pairs); err != nil {
		return nil, err
	}

	if pairs == nil || len(pairs) > 1024 {
		return nil, errors.New("telegram retired identity interval capacity exceeded")
	}

	result := make([]retiredInterval, 0, len(pairs))
	for _, pair := range pairs {
		if !validRetiredInterval(result, pair) {
			return nil, errors.New("invalid telegram retired identity intervals")
		}

		result = append(result, retiredInterval{pair[0], pair[1]})
	}

	return result, nil
}

func validRetiredInterval(prior []retiredInterval, pair []int64) bool {
	if len(pair) != 2 || pair[0] < 0 || pair[1] < pair[0] || pair[1] == math.MaxInt64 {
		return false
	}

	return len(prior) == 0 || pair[0] > prior[len(prior)-1][1]+1
}

func updateRetired(intervals []retiredInterval, id int64) bool {
	index := sort.Search(len(intervals), func(i int) bool { return intervals[i][1] >= id })

	return index < len(intervals) && intervals[index][0] <= id
}

func addRetired(intervals []retiredInterval, id int64) []retiredInterval {
	sorted := append(append([]retiredInterval{}, intervals...), retiredInterval{id, id})
	sort.Slice(sorted, func(i, j int) bool { return sorted[i][0] < sorted[j][0] })

	result := make([]retiredInterval, 0, len(sorted))
	for _, next := range sorted {
		if len(result) > 0 && next[0] <= result[len(result)-1][1]+1 {
			result[len(result)-1][1] = max(result[len(result)-1][1], next[1])
		} else {
			result = append(result, next)
		}
	}

	return result
}
