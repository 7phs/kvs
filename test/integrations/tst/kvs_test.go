package tst

import (
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type KvsSuite struct {
	suite.Suite

	client Client
}

func (suite *KvsSuite) TestAdd() {
	key := randomString()
	value := randomString()

	err := suite.client.Add(key, value)
	suite.Require().NoError(err)

	storedValue, err := suite.client.Get(key)
	suite.Require().NoError(err)

	suite.Equal(value, storedValue)
}

func (suite *KvsSuite) TestGetTwice() {
	key := randomString()
	value := randomString()

	err := suite.client.Add(key, value)
	suite.Require().NoError(err)

	for i := 0; i < 2; i++ {
		storedValue, err := suite.client.Get(key)
		suite.Require().NoError(err)
		suite.Equal(value, storedValue)
	}
}

func (suite *KvsSuite) TestGetNotFound() {
	key := randomString()

	storedValue, err := suite.client.Get(key)
	suite.Require().Error(err)
	suite.EqualError(err, ErrNotFound.Error())
	suite.Empty(storedValue)
}

func randomString() string {
	return strconv.FormatInt(1000000+rand.Int63(), 16)
}

func TestKvs(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	s := &KvsSuite{
		client: NewClient(os.Getenv("KVS")),
	}

	suite.Run(t, s)
}
