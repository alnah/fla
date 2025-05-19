package pathutil

import "os"

const (
	// User permissions
	PermUserRWX os.FileMode = 0o700
	PermUserRW  os.FileMode = 0o600
	PermUserR   os.FileMode = 0o400
)
