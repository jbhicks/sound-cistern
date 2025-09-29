package contract

import (
	"testing"

	"github.com/gobuffalo/suite/v4"
	"github.com/jbhicks/sound-cistern/actions"
)

type AuthSoundcloudContractSuite struct {
	*suite.Action
}

func Test_AuthSoundcloudContract(t *testing.T) {
	as := &AuthSoundcloudContractSuite{
		Action: suite.NewAction(actions.App()),
	}
	suite.Run(t, as)
}

func (as *AuthSoundcloudContractSuite) Test_AuthSoundcloud() {
	res := as.HTML("/auth/soundcloud").Get()
	// Expect redirect to Soundcloud (302)
	as.Equal(302, res.Code)
}
