package core

import (
	"os"
	"strconv"
	"strings"

	"github.com/m0rk0vka/passive_investing/pkg/telegram/services/poller"
)

var _ poller.OffsetKeepper = (*offsetKeepper)(nil)

type offsetKeepper struct {
}

func NewOffsetKeepper() *offsetKeepper {
	return &offsetKeepper{}
}

func (o *offsetKeepper) GetOffset() (int, error) {
	b, err := os.ReadFile(".offset")
	if err != nil {
		return 0, err
	}
	n, err := strconv.Atoi(strings.TrimSpace(string(b)))
	if err != nil {
		return 0, err
	}
	return n, nil
}

func (o *offsetKeepper) SetOffset(offset int) error {
	return os.WriteFile(".offset", []byte(strconv.Itoa(offset)), 0o644)
}
