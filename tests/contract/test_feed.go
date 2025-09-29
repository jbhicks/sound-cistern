package contract

import (
	"testing"

	"github.com/gobuffalo/suite/v4"
	"github.com/jbhicks/sound-cistern/actions"
)

type FeedContractSuite struct {
	*suite.Action
}

func Test_FeedContract(t *testing.T) {
	as := &FeedContractSuite{
		Action: suite.NewAction(actions.App()),
	}
	suite.Run(t, as)
}

func (as *FeedContractSuite) Test_Feed() {
	// TODO: Add authentication setup first
	// res := as.HTML("/feed").Get()
	// as.Equal(200, res.Code)
	// Expect list of tracks
}
