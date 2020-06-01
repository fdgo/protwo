package service

import (
	"errors"
	"fmt"
	"github.com/wangmhgo/go-project/gdun/common"
	"math/rand"
	"strconv"
	"time"
)

const (
	MAX_INVITATION_TOKEN = 10
)

//----------------------------------------------------------------------------

func (us *UserService) GenerateInvitationToken(size int, groupID int) ([]string, error) {
	if size <= 0 || size > MAX_INVITATION_TOKEN {
		return nil, errors.New("Invalid size.")
	}

	if groupID <= 1 {
		return nil, errors.New("Invalid group ID.")
	}

	r := make([]string, size)
	for i := 0; i < size; i++ {
		token := fmt.Sprintf("%016x%016x", time.Now().UnixNano(), rand.Int())
		us.cache.SetField(common.FIELD_INVITATION, token, strconv.Itoa(groupID))
		r[i] = token
	}

	return r, nil
}

//----------------------------------------------------------------------------

func (us *UserService) QueryInvitationToken() (map[string]string, error) {
	return us.cache.GetAllFields(common.FIELD_INVITATION)
}

//----------------------------------------------------------------------------

func (us *UserService) DeleteInvitationToken(token string) error {
	if len(token) == 0 {
		return errors.New("Empty token.")
	}

	return us.cache.DelField(common.FIELD_INVITATION, token)
}

//----------------------------------------------------------------------------

func (us *UserService) UseInvitationToken(token string) (int, error) {
	if len(token) == 0 {
		return 0, errors.New("Empty token.")
	}

	// Get this key from cache.
	s, err := us.cache.GetField(common.FIELD_INVITATION, token)
	if err != nil {
		return 0, err
	}

	// Get the group ID assigned to this token.
	id, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}

	// Delete this invitation token from cache.
	err = us.cache.DelField(common.FIELD_INVITATION, token)
	if err != nil {
		return id, err
	}

	return id, nil
}

//----------------------------------------------------------------------------
