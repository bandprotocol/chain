package priceservice

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

const (
	flagQueryTimeout = "timeout"
)

var (
	ErrExecutionimeout = errors.New("execution timeout")
	ErrRestNotOk       = errors.New("rest return non 2XX response")
)

type PriceService interface {
	Query(symbols []string) ([]types.SubmitPrice, error)
}

// NewPriceService returns priceService by name and priceService URL
func PriceServiceFromUrl(priceService string) (exec PriceService, err error) {
	name, base, timeout, err := parsePriceServiceURL(priceService)
	if err != nil {
		return nil, err
	}
	switch name {
	case "rest":
		exec = NewRestService(base, timeout)
	case "grpc":
		exec = NewGRPCService(base, timeout)
	default:
		return nil, fmt.Errorf("invalid priceService name: %s, base: %s", name, base)
	}

	// TODO: Remove hardcode in test execution
	_, err = exec.Query([]string{"BTC"})
	if err != nil {
		return nil, fmt.Errorf("failed to run test program: %s", err.Error())
	}

	return exec, nil
}

// parsePriceService splits the priceService string in the form of "name:base?timeout=" into parts.
func parsePriceServiceURL(priceServiceStr string) (name string, base string, timeout time.Duration, err error) {
	priceService := strings.SplitN(priceServiceStr, ":", 2)
	if len(priceService) != 2 {
		return "", "", 0, fmt.Errorf("invalid priceService, cannot parse priceService: %s", priceServiceStr)
	}
	u, err := url.Parse(priceService[1])
	if err != nil {
		return "", "", 0, fmt.Errorf(
			"invalid url, cannot parse %s to url with error: %s",
			priceService[1],
			err.Error(),
		)
	}

	query := u.Query()
	timeoutStr := query.Get(flagQueryTimeout)
	if timeoutStr == "" {
		return "", "", 0, fmt.Errorf("invalid timeout, priceService requires query timeout")
	}
	// Remove timeout from query because we need to return `base`
	query.Del(flagQueryTimeout)
	u.RawQuery = query.Encode()

	timeout, err = time.ParseDuration(timeoutStr)
	if err != nil {
		return "", "", 0, fmt.Errorf("invalid timeout, cannot parse duration with error: %s", err.Error())
	}
	return priceService[0], u.String(), timeout, nil
}
