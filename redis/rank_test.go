package redis

import (
	"strconv"
	"testing"

	"github.com/laonsx/gamelib/codec"
)

func init() {

	InitRedis(codec.MsgPack, codec.UnMsgPack, NewRedisConf("rank", "127.0.0.1", "6378", 0))
}

type TestRank struct {
}

func (rank *TestRank) Name() string {

	return "test"
}

func (rank *TestRank) Key() string {

	return "test"
}

func TestRank_Set_Get_Incrby(t *testing.T) {

	testRank := NewRank(&TestRank{})

	for i := 1; i < 2000; i++ {

		err := testRank.SetRankScore(testRank.GetKey("II"), i, i)
		if err != nil {

			t.Error("SetRankScore", "II", i, err)
		}
	}

	for j := 2000; j < 4000; j++ {

		err := testRank.SetRankScore(testRank.GetKey("IV"), j, j)
		if err != nil {

			t.Error("SetRankScore", "IV", j, err)
		}
	}

	rankList, err := testRank.GetRankByPage(testRank.GetKey("II"), 1, 10)
	if err != nil {

		t.Error("GetRankByPage", "II", 1, 10, err)
	}

	for index, rankData := range rankList {
		uid, _ := strconv.ParseUint(rankData[0], 10, 64)
		score, _ := strconv.ParseInt(rankData[1], 10, 64)

		t.Log("II", "index:", index, "uid:", uid, "score:", score)
	}

	score, err := testRank.IncrbyRankScore(testRank.GetKey("IV"), 2000, 2000)
	if err != nil {

		t.Error("IncrbyRankScore", "IV", 2000, 2000, err, score)
	}

	t.Log(2000, score)

	rankList, err = testRank.GetRankByPage(testRank.GetKey("IV"), 1, 10)
	if err != nil {

		t.Error("GetRankByPage", "IV", 1, 10, err)
	}

	for index, rankData := range rankList {
		uid, _ := strconv.ParseUint(rankData[0], 10, 64)
		score, _ := strconv.ParseInt(rankData[1], 10, 64)

		t.Log("IV", "index:", index, "uid:", uid, "score:", score)
	}
}
