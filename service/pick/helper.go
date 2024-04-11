package pick

import (
	"crypto/md5"
	"fmt"
)

func sumMd5(s string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}
