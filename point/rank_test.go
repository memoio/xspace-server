package point

import (
	"math/rand"
	"sort"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/memoio/xspace-server/types"
)

func TestRank(t *testing.T) {
	projects, err := ListProjects()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(projects)

	for _, project := range projects {
		rankInfo, n, err := GetRank(project.ProjectID, 1, 10)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(n)
		t.Log(rankInfo)
		t.Log("")

		rankInfo, _, err = GetRank(project.ProjectID, 50, 10)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(rankInfo)
		t.Log("")

		rankInfo, _, err = GetRank(project.ProjectID, 200, 10)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(rankInfo)
		t.Log("")
		t.Log("--------------------")
	}
}

func generateRandomRankData(count int) []types.RankInfo {
	var result []types.RankInfo

	for i := 1; i <= count; i++ {
		// 生成20~9800之间的10的整数倍分数
		score := (rand.Intn(1203) + 2) * 10

		sk, err := crypto.GenerateKey()
		if err != nil {
			return nil
		}

		result = append(result, types.RankInfo{
			Rank:    i,
			Address: crypto.PubkeyToAddress(sk.PublicKey).Hex(),
			Scores:  int64(score),
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Scores > result[j].Scores
	})

	// 更新排名字段(根据排序后的位置)
	for i := range result {
		result[i].Rank = i + 1
		result[i].Points = getRankPoint(i + 1)
	}

	return result
}
