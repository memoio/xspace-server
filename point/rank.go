package point

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/memoio/xspace-server/types"
	"golang.org/x/xerrors"
)

var dataDIDStart = time.Date(2025, time.April, 1, 0, 0, 0, 0, time.Local)
var metisStart = time.Date(2025, time.March, 5, 0, 0, 0, 0, time.Local)
var arkreenStart = time.Date(2025, time.March, 20, 0, 0, 0, 0, time.Local)
var adotStart = time.Date(2025, time.March, 28, 0, 0, 0, 0, time.Local)

var metisInfo []types.RankInfo = nil
var arkreenInfo []types.RankInfo = nil
var adotInfo []types.RankInfo = nil

var defaultProjects = map[int]ProjectInfo{
	1: {
		ID:       1,
		Name:     "Data DID",
		Start:    dataDIDStart,
		End:      dataDIDStart.Add(7 * 24 * time.Hour),
		RankFunc: getDIDRank,
	},
	2: {
		ID:       2,
		Name:     "Metis",
		Start:    metisStart,
		End:      metisStart.Add(7 * 24 * time.Hour),
		RankFunc: getMetisRank,
	},
	3: {
		ID:       3,
		Name:     "arkreen",
		Start:    arkreenStart,
		End:      arkreenStart.Add(7 * 24 * time.Hour),
		RankFunc: getArkreenRank,
	},
	4: {
		ID:       4,
		Name:     "Adot",
		Start:    adotStart,
		End:      adotStart.Add(7 * 24 * time.Hour),
		RankFunc: getAdotRank,
	},
}

type rankFunc func(int, int) ([]types.RankInfo, int, error)

type ProjectInfo struct {
	ID       int
	Name     string
	Start    time.Time
	End      time.Time
	RankFunc rankFunc
}

func init() {
	content, err := os.ReadFile("./adot.json")
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(content, &adotInfo)
	if err != nil {
		panic(err)
	}

	content, err = os.ReadFile("./arkreen.json")
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(content, &arkreenInfo)
	if err != nil {
		panic(err)
	}

	content, err = os.ReadFile("./metis.json")
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(content, &metisInfo)
	if err != nil {
		panic(err)
	}
}

func ListProjects() ([]types.ProjectInfo, error) {
	var projects []types.ProjectInfo
	for _, project := range defaultProjects {
		projects = append(projects, types.ProjectInfo{
			ProjectID: project.ID,
			Name:      project.Name,
			Start:     project.Start,
			End:       project.End,
		})
	}

	sort.Slice(projects, func(i, j int) bool {
		return projects[i].ProjectID < projects[j].ProjectID
	})

	return projects, nil
}

func GetRank(id, page, size int) ([]types.RankInfo, int, error) {
	project, ok := defaultProjects[id]
	if !ok {
		return nil, 0, xerrors.Errorf("Unspported project id %d", id)
	}

	return project.RankFunc(page, size)
}

func getDIDRank(page, size int) ([]types.RankInfo, int, error) {
	client := &http.Client{Timeout: time.Minute}
	var url = "https://data-be.metamemo.one/airdrop/rank"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, 0, err
	}

	params := req.URL.Query()
	params.Add("type", "0")
	req.URL.RawQuery = params.Encode()

	res, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, 0, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, 0, xerrors.Errorf(string(data))
	}

	var rankInfo = struct {
		Result int
		Data   []struct {
			Uid           string
			Points        int64
			NickName      string
			WalletAddress string
			Avatar        string
			Inviter       string
			InviteCount   int
		}
		Error string
	}{}
	var result []types.RankInfo = make([]types.RankInfo, size)

	err = json.Unmarshal(data, &rankInfo)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * size
	for index := 0; index < size; index++ {
		if offset+index >= len(rankInfo.Data) {
			return result[:index], 0, nil
		}

		result[index] = types.RankInfo{
			Rank:    offset + index + 1,
			Address: rankInfo.Data[offset+index].WalletAddress,
			Scores:  rankInfo.Data[offset+index].Points,
			Points:  getRankPoint(offset + index + 1),
		}
	}

	return result, len(rankInfo.Data), nil
}

func getMetisRank(page, size int) ([]types.RankInfo, int, error) {
	if metisInfo == nil {
		return nil, 0, xerrors.Errorf("No metis rank data")
	}

	if (page-1)*size > len(metisInfo) {
		return nil, len(metisInfo), nil
	}

	if page*size > len(metisInfo) {
		return metisInfo[(page-1)*size:], len(metisInfo), nil
	}

	return metisInfo[(page-1)*size : page*size], len(metisInfo), nil
}

func getArkreenRank(page, size int) ([]types.RankInfo, int, error) {
	if arkreenInfo == nil {
		return nil, 0, xerrors.Errorf("No arkreen rank data")
	}

	if (page-1)*size > len(arkreenInfo) {
		return nil, len(arkreenInfo), nil
	}

	if page*size > len(arkreenInfo) {
		return arkreenInfo[(page-1)*size:], len(arkreenInfo), nil
	}

	return arkreenInfo[(page-1)*size : page*size], len(arkreenInfo), nil
}

func getAdotRank(page, size int) ([]types.RankInfo, int, error) {
	if adotInfo == nil {
		return nil, 0, xerrors.Errorf("No adot rank data")
	}

	if (page-1)*size > len(adotInfo) {
		return nil, len(adotInfo), nil
	}

	if page*size > len(adotInfo) {
		return adotInfo[(page-1)*size:], len(adotInfo), nil
	}

	return adotInfo[(page-1)*size : page*size], len(adotInfo), nil
}

func getRankPoint(rank int) int64 {
	if rank <= 0 {
		return 0
	}

	switch rank {
	case 1:
		return 10000
	case 2:
		return 8000
	case 3:
		return 6000
	case 4:
		return 5000
	case 5:
		return 4000
	case 6:
		return 3800
	case 7:
		return 3600
	case 8:
		return 3400
	case 9:
		return 3200
	case 10:
		return 3000
	}

	if rank <= 20 {
		return 2000
	} else if rank <= 50 {
		return 1000
	} else if rank <= 100 {
		return 800
	} else if rank <= 200 {
		return 400
	} else if rank <= 300 {
		return 200
	} else if rank <= 500 {
		return 100
	} else if rank <= 1000 {
		return 60
	}

	return 0
}
