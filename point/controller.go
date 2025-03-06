package point

import (
	"time"

	"github.com/memoio/xspace-server/database"
	"golang.org/x/xerrors"
)

var defaultActions = map[int]ActionInfo{
	1: {
		ID:          1,
		Name:        "Sigin",
		Description: "Log in to xspace and sign in every day",
		ResetTime:   24 * time.Hour,
		Point:       5,
	},
	2: {
		ID:          2,
		Name:        "Charging",
		Description: "Daily charging",
		ResetTime:   5 * time.Hour,
		Point:       3,
	},
	3: {
		ID:          3,
		Name:        "MintNFT",
		Description: "Mint tweets into NFT",
		ResetTime:   time.Duration(0),
		Point:       15,
	},

	11: {
		ID:          11,
		Name:        "Invited",
		Description: "Invited by other xspace's user",
		ResetTime:   24 * time.Hour,
		Point:       50,
	},
	12: {
		ID:          12,
		Name:        "Invite",
		Description: "Invited a new user",
		ResetTime:   time.Duration(0),
		Point:       50,
	},

	100: {
		ID:          100,
		Name:        "DaliyCheckIn",
		Description: "Visit MEMO's Twitter daily",
		ResetTime:   24 * time.Hour,
		Point:       5,
	},
	101: {
		ID:          101,
		Name:        "FollowTwitter",
		Description: "Follow MEMO's official account on Twitter",
		ResetTime:   time.Duration(-1),
		Point:       50,
	},
	102: {
		ID:          102,
		Name:        "FollowDiscord",
		Description: "Follow MEMO's official account on Discord",
		ResetTime:   time.Duration(-1),
		Point:       50,
	},
	103: {
		ID:          103,
		Name:        "FollowTelegram",
		Description: "Follow MEMO's official account on Telegram",
		ResetTime:   time.Duration(-1),
		Point:       50,
	},
	104: {
		ID:          104,
		Name:        "ShareToTelegram",
		Description: "Share the invite link to the Telegram group",
		ResetTime:   24 * time.Hour,
		Point:       5,
	},
	105: {
		ID:          105,
		Name:        "ShareToTwitter",
		Description: "Share the invite link to the Twitter",
		ResetTime:   24 * time.Hour,
		Point:       5,
	},
}

type ActionInfo struct {
	ID          int
	Name        string
	Description string
	ResetTime   time.Duration
	Point       int64
}

type PointController struct {
	Actions map[int]ActionInfo
}

func NewPointController() (*PointController, error) {
	return &PointController{Actions: defaultActions}, nil
}

func (c *PointController) GetActionInfo(actionID int) (ActionInfo, error) {
	actionInfo, ok := c.Actions[actionID]
	if !ok {
		return actionInfo, xerrors.Errorf("Unsupported action id: %d", actionID)
	}

	return actionInfo, nil
}

func (c *PointController) FinishAction(address string, actionID int) (database.UserStore, error) {
	if actionID == 11 || actionID == 12 {
		return database.UserStore{}, xerrors.New("not support refer action, please use FinishInvited function")
	}

	actionInfo, err := c.GetActionInfo(actionID)
	if err != nil {
		return database.UserStore{}, err
	}

	err = c.checkExpire(address, actionInfo)
	if err != nil {
		return database.UserStore{}, err
	}

	userInfo, err := database.GetUserInfo(address)
	if err != nil {
		return database.UserStore{}, err
	}

	userInfo.Points += actionInfo.Point
	if actionID == 3 {
		if userInfo.Space == 0 {
			return database.UserStore{}, xerrors.New("The user's current storage units is 0")
		}
		userInfo.Space -= 1
	}

	err = userInfo.UpdateUserInfo()
	if err != nil {
		return database.UserStore{}, err
	}

	action := database.ActionStore{
		ActionId: actionInfo.ID,
		Name:     actionInfo.Name,
		Address:  address,
		Point:    actionInfo.Point,
		Time:     time.Now(),
	}

	return userInfo, action.CreateActionInfo()
}

func (c *PointController) checkExpire(address string, actionInfo ActionInfo) error {
	actions, err := database.ListActionHistoryByID(address, 1, 5, "date_desc", actionInfo.ID)
	if err != nil {
		return err
	}

	if len(actions) > 0 {
		if actionInfo.ResetTime == -1 {
			return xerrors.Errorf("%s is one-time action", actionInfo.Name)
		}
		if actions[0].Time.Add(actionInfo.ResetTime).After(time.Now()) {
			return xerrors.Errorf("The last %s time is %s, please try again after %s", actionInfo.Name, actions[0].Time.String(), actions[0].Time.Add(actionInfo.ResetTime).String())
		}
	}

	return nil
}
