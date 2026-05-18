package helpers

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/xo/dburl"

	"github.com/itisbryan/oh-my-lazysql/commands"
)

func ParseConnectionString(url string) (*dburl.URL, error) {
	return dburl.Parse(url)
}

func BuildConnectionURL(provider, username, password, hostname, port, dbname string, sslEnabled bool, sslCert, sslKey, sslCA string) (string, error) {
	switch provider {
	case "mysql":
		url := fmt.Sprintf("mysql://%s:%s@%s:%s/%s", url.QueryEscape(username), url.QueryEscape(password), hostname, port, dbname)
		if sslEnabled {
			url += "?tls=true"
			if sslCert != "" || sslKey != "" || sslCA != "" {
				url += "&tls=true"
			}
		}
		return url, nil

	case "postgres":
		url := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", url.QueryEscape(username), url.QueryEscape(password), hostname, port, dbname)
		if sslEnabled {
			url += "?sslmode=require"
		}
		return url, nil

	case "sqlite3":
		if dbname == "" {
			return "", errors.New("database name required for SQLite")
		}
		return fmt.Sprintf("sqlite3:%s", dbname), nil

	case "sqlserver":
		url := fmt.Sprintf("sqlserver://%s:%s@%s:%s?database=%s", url.QueryEscape(username), url.QueryEscape(password), hostname, port, dbname)
		return url, nil

	default:
		return "", errors.New("unsupported provider: " + provider)
	}
}

func ContainsCommand(commands []commands.Command, command commands.Command) bool {
	for _, cmd := range commands {
		if cmd == command {
			return true
		}
	}
	return false
}

// GetFreePort asks the kernel for a free port.
func GetFreePort() (string, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return "", err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return "", err
	}
	defer l.Close()

	return strconv.Itoa(l.Addr().(*net.TCPAddr).Port), nil
}

// WaitForPort waits for a port to be open.
func WaitForPort(ctx context.Context, port string) error {
	dialer := &net.Dialer{
		Timeout: 500 * time.Millisecond,
	}

	for i := 0; i < 10; i++ {
		conn, err := dialer.DialContext(ctx, "tcp", "localhost:"+port)
		if err == nil {
			_ = conn.Close()
			return nil
		}

		select {
		case <-time.After(500 * time.Millisecond):
			continue
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return errors.New("Timeout waiting for port " + port)
}
