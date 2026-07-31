package backup

import (
	"bytes"
	"os"
	"os/exec"
)

// pgDump runs pg_dump to create a PostgreSQL backup file.
func pgDump(dbc dbConfig, filepath string) error {
	cmd := exec.Command(findExe("pg_dump"),
		"-h", dbc.Host,
		"-p", dbc.Port,
		"-U", dbc.User,
		"-d", dbc.DBName,
		"--no-owner", "--no-acl",
		"-f", filepath,
	)
	cmd.Env = append(os.Environ(), "PGPASSWORD="+dbc.Password, "PGSSLMODE=disable")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return &pgError{msg: string(output)}
	}
	return nil
}

// pgRestore pipes SQL through psql to restore a PostgreSQL backup.
func pgRestore(dbc dbConfig, content []byte) error {
	cmd := exec.Command(findExe("psql"),
		"-h", dbc.Host,
		"-p", dbc.Port,
		"-U", dbc.User,
		"-d", dbc.DBName,
	)
	cmd.Env = append(os.Environ(), "PGPASSWORD="+dbc.Password, "PGSSLMODE=disable")
	cmd.Stdin = bytes.NewReader(content)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return &pgError{msg: string(output)}
	}
	return nil
}

type pgError struct{ msg string }

func (e *pgError) Error() string { return e.msg }

// findExe searches for an executable in common paths.
func findExe(name string) string {
	paths := []string{
		"/opt/homebrew/bin/" + name,
		"/usr/local/bin/" + name,
		"/usr/bin/" + name,
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	if p, err := exec.LookPath(name); err == nil {
		return p
	}
	return name
}
